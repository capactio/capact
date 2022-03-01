import { Transaction } from "neo4j-driver";
import { ContextWithDriver } from "./context";
import { tryToExtractCustomError } from "./helpers";
import { UpdateTypeInstanceErrorCode } from "./update-type-instances";

export async function deleteTypeInstance(
  _: any,
  args: { id: string; ownerID: string },
  context: ContextWithDriver
) {
  const neo4jSession = context.driver.session();
  try {
    return await neo4jSession.writeTransaction(async (tx: Transaction) => {
      await tx.run(
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
                    MATCH (metadata:TypeInstanceResourceVersionMetadata)<-[:DESCRIBED_BY]-(tirs)
                    MATCH (tirs)-[:SPECIFIED_BY]->(spec: TypeInstanceResourceVersionSpec)
                    OPTIONAL MATCH (metadata)-[:CHARACTERIZED_BY]->(attrRef: AttributeReference)

                    DETACH DELETE ti, metadata, spec, tirs

                    WITH typeRef
                    CALL {
                      MATCH (typeRef)
                      WHERE NOT (typeRef)--()
                      DELETE (typeRef)
                      RETURN 'remove typeRef'
                    }

                    WITH *
                    CALL {
                      MATCH (attrRef)
                      WHERE attrRef IS NOT NULL AND NOT (attrRef)--()
                      DELETE (attrRef)
                      RETURN 'remove attr'
                    }

                    RETURN $id`,
        { id: args.id, ownerID: args.ownerID || null }
      );
      return args.id;
    });
  } catch (e) {
    let err = e as Error;
    const customErr = tryToExtractCustomError(err);
    if (customErr) {
      switch (customErr.code) {
        case UpdateTypeInstanceErrorCode.Conflict:
          err = Error(`TypeInstance is locked by different owner`);
          break;
        case UpdateTypeInstanceErrorCode.NotFound:
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
