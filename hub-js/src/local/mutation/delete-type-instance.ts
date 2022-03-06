import { Transaction } from "neo4j-driver";
import { Context } from "./context";
import {
  CustomCypherErrorCode,
  tryToExtractCustomCypherError,
} from "./cypher-errors";
import { logger } from "../../logger";

export async function deleteTypeInstance(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  _: any,
  args: { id: string; ownerID: string },
  context: Context
) {
  const neo4jSession = context.driver.session();
  try {
    return await neo4jSession.writeTransaction(async (tx: Transaction) => {
      logger.debug(
        "Executing query to delete TypeInstance from database",
        args
      );
      const result = await tx.run(
        `
            OPTIONAL MATCH (ti:TypeInstance {id: $id})

            // Check if a given TypeInstance was found
            CALL apoc.util.validate(ti IS NULL, apoc.convert.toJson({code: 404}), null)

            // Check if a given TypeInstance is not already locked by a different owner
            CALL {
                WITH ti
                WITH ti
                WHERE ti.lockedBy IS NOT NULL AND ($ownerID IS NULL OR ti.lockedBy <> $ownerID)
                WITH count(ti.id) as lockedIDs
                RETURN lockedIDs = 1 as isLocked
            }
            CALL apoc.util.validate(isLocked, apoc.convert.toJson({code: 409}), null)

            // Check if a given TypeInstance is not used by others
            CALL {
                WITH ti
                WITH ti
                MATCH (ti)-[:USES]->(others:TypeInstance)
                WITH count(others) as othersLen
                RETURN  othersLen > 1 as isUsed
            }
            CALL apoc.util.validate(isUsed, apoc.convert.toJson({code: 400}), null)

            WITH ti
            MATCH (ti)-[:CONTAINS]->(tirs: TypeInstanceResourceVersion)
            MATCH (ti)-[:OF_TYPE]->(typeRef: TypeInstanceTypeReference)
            MATCH (ti)-[:STORED_IN]->(backendRef: TypeInstanceBackendReference)
            MATCH (metadata:TypeInstanceResourceVersionMetadata)<-[:DESCRIBED_BY]-(tirs)
            MATCH (tirs)-[:SPECIFIED_BY]->(spec: TypeInstanceResourceVersionSpec)
            MATCH (spec)-[:WITH_BACKEND]->(specBackend: TypeInstanceResourceVersionSpecBackend)

            OPTIONAL MATCH (metadata)-[:CHARACTERIZED_BY]->(attrRef: AttributeReference)

            // NOTE: Need to be preserved with 'WITH' statement, otherwise we won't be able
            // to access node's properties after 'DETACH DELETE' statement.
            WITH *, {id: ti.id, backend: { id: backendRef.id, context: specBackend.context, abstract: backendRef.abstract}} as out
            DETACH DELETE ti, metadata, spec, tirs, specBackend

            WITH *
            CALL {
              MATCH (typeRef)
              WHERE NOT (typeRef)--()
              DELETE (typeRef)
              RETURN 'remove typeRef'
            }

            WITH *
            CALL {
              MATCH (backendRef)
              WHERE NOT (backendRef)--()
              DELETE (backendRef)
              RETURN 'remove backendRef'
            }

            WITH *
            CALL {
              OPTIONAL MATCH (attrRef)
              WHERE attrRef IS NOT NULL AND NOT (attrRef)--()
              DELETE (attrRef)
              RETURN 'remove attr'
            }

            RETURN out`,
        { id: args.id, ownerID: args.ownerID || null }
      );

      const deleteExternally = new Map<string, any>();
      result.records.forEach((record) => {
        const out = record.get("out");
        if (out.backend.abstract) {
          return;
        }
        deleteExternally.set(out.id, out.backend);
      });

      for (const [id, backend] of deleteExternally) {
        await context.delegatedStorage.Delete({
          typeInstance: {
            id,
          },
          backend,
        });
      }

      return args.id;
    });
  } catch (e) {
    let err = e as Error;
    const customErr = tryToExtractCustomCypherError(err);
    if (customErr) {
      switch (customErr.code) {
        case CustomCypherErrorCode.Conflict:
          err = Error(`TypeInstance is locked by different owner`);
          break;
        case CustomCypherErrorCode.NotFound:
          err = Error(`TypeInstance was not found`);
          break;
        default:
          err = Error(`Unexpected error code ${customErr.code}`);
          break;
      }
    }

    throw new Error(
      `failed to delete TypeInstance with ID "${args.id}": ${err.message}`
    );
  } finally {
    await neo4jSession.close();
  }
}