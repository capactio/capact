import { readFileSync } from "fs";
import { makeAugmentedSchema, neo4jgraphql } from "neo4j-graphql-js";
import { Driver, Transaction } from "neo4j-driver";

const typeDefs = readFileSync("./graphql/local/schema.graphql", "utf-8");

interface TypeInstanceNode {
  properties: { id: string; lockedBy: string };
}

interface CreateTypeInstancesArgs {
  in: {
    typeInstances: Array<{ alias: string }>;
    usesRelations: Array<{ from: string; to: string }>;
  };
}

interface ContextWithDriver {
  driver: Driver;
}

interface LockingTypeInstanceInput {
  in: {
    ids: [string];
    ownerID: string;
  };
}

interface LockingResult {
  allIDs: [TypeInstanceNode];
  lockedIDs: [TypeInstanceNode];
  lockingProcess: {
    executed: boolean;
  };
}

enum UpdateTypeInstanceErrorCode {
  Conflict = 409,
  NotFound = 404,
}

interface UpdateTypeInstanceError {
  code: UpdateTypeInstanceErrorCode;
  ids: string[];
}

// TODO: extract each mutation/query into dedicated file
export const schema = makeAugmentedSchema({
  typeDefs,
  resolvers: {
    Mutation: {
      createTypeInstances: async (
        _: any,
        args: CreateTypeInstancesArgs,
        context: ContextWithDriver
      ) => {
        const { typeInstances, usesRelations } = args.in;

        const aliases = typeInstances
          .filter((x) => x.alias !== undefined)
          .map((x) => x.alias);
        if (new Set(aliases).size !== aliases.length) {
          throw new Error(
            "Failed to create TypeInstances, due to duplicated TypeInstance aliases. Please ensure that each TypeInstance alias is unique."
          );
        }

        const neo4jSession = context.driver.session();

        try {
          return await neo4jSession.writeTransaction(
            async (tx: Transaction) => {
              // create TypeInstances
              const createTypeInstanceResult = await tx.run(
                `UNWIND $typeInstances AS typeInstance
               MERGE (typeRef:TypeInstanceTypeReference {path: typeInstance.typeRef.path, revision: typeInstance.typeRef.revision})
               
               CREATE (ti:TypeInstance {id: apoc.create.uuid()})
               CREATE (ti)-[:OF_TYPE]->(typeRef)
               
               CREATE (tir: TypeInstanceResourceVersion {resourceVersion: 1, createdBy: typeInstance.createdBy})
               CREATE (ti)-[:CONTAINS]->(tir)
               
               CREATE (tir)-[:DESCRIBED_BY]->(metadata: TypeInstanceResourceVersionMetadata)
               CREATE (tir)-[:SPECIFIED_BY]->(spec: TypeInstanceResourceVersionSpec {value: apoc.convert.toJson(typeInstance.value)})

               FOREACH (attr in typeInstance.attributes |
                 MERGE (attrRef: AttributeReference {path: attr.path, revision: attr.revision})
                 CREATE (metadata)-[:CHARACTERIZED_BY]->(attrRef)
               )

               RETURN ti.id as uuid, typeInstance.alias as alias
              `,
                { typeInstances }
              );

              if (
                createTypeInstanceResult.records.length !== typeInstances.length
              ) {
                throw new Error(
                  "Failed to create some TypeInstances. Please verify, if you provided all the required fields for TypeInstances."
                );
              }

              const aliasMappings: {
                [key: string]: string;
              } = createTypeInstanceResult.records.reduce(
                (acc: { [key: string]: string }, cur) => {
                  const uuid = cur.get("uuid");
                  const alias = cur.get("alias");

                  return {
                    ...acc,
                    [alias || uuid]: uuid,
                  };
                },
                {}
              );
              const usesRelationsParams = usesRelations.map(
                ({ from, to }: { from: string; to: string }) => ({
                  from: aliasMappings[from] || from,
                  to: aliasMappings[to] || to,
                })
              );

              const createRelationsResult = await tx.run(
                `UNWIND $usesRelations as usesRelation
               MATCH (fromTi:TypeInstance {id: usesRelation.from})
               MATCH (toTi:TypeInstance {id: usesRelation.to})
               CREATE (fromTi)-[:USES]->(toTi)
               RETURN fromTi AS from, toTi AS to
              `,
                {
                  usesRelations: usesRelationsParams,
                }
              );

              if (
                createRelationsResult.records.length !==
                usesRelationsParams.length
              ) {
                throw new Error(
                  "Failed to create some relations. Please verify, if you use proper aliases or IDs in relations definition."
                );
              }

              return Object.entries(aliasMappings).map((entry) => ({
                alias: entry[0],
                id: entry[1],
              }));
            }
          );
        } catch (e) {
          throw new Error(`failed to create the TypeInstances: ${e.message}`);
        } finally {
          await neo4jSession.close();
        }
      },
      updateTypeInstances: async (
        obj,
        args: { in: [{ id: string }] },
        context,
        resolveInfo
      ) => {
        try {
          return await neo4jgraphql(obj, args, context, resolveInfo);
        } catch (e) {
          let err = e;
          const customErr = tryToExtractCustomError(err);
          if (customErr) {
            switch (customErr.code) {
              case UpdateTypeInstanceErrorCode.Conflict:
                err = Error(
                  `TypeInstances with IDs "${customErr.ids.join(
                    '", "'
                  )}" are locked by different owner`
                );
                break;
              case UpdateTypeInstanceErrorCode.NotFound: {
                const ids = args.in.map(({ id }) => id);
                const notFoundIDs = ids.filter(
                  (x) => !customErr.ids.includes(x)
                );
                err = Error(
                  `TypeInstances with IDs "${notFoundIDs.join(
                    '", "'
                  )}" were not found`
                );
                break;
              }
              default:
                err = Error(`Unexpected error code ${customErr.code}`);
                break;
            }
          }

          throw new Error(`failed to update TypeInstances: ${err.message}`);
        }
      },
      deleteTypeInstance: async (_obj, args, context) => {
        const neo4jSession = context.driver.session();
        try {
          return await neo4jSession.writeTransaction(
            async (tx: Transaction) => {
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
            }
          );
        } catch (e) {
          let err = e;
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
          neo4jSession.close();
        }
      },
      lockTypeInstances: async (
        _obj,
        args: LockingTypeInstanceInput,
        context
      ) => {
        const neo4jSession = context.driver.session();
        try {
          return await neo4jSession.writeTransaction(
            async (tx: Transaction) => {
              await switchLocking(
                tx,
                args,
                `
                  MATCH (ti:TypeInstance)
                  WHERE ti.id IN $in.ids
                  SET ti.lockedBy = $in.ownerID
                  RETURN true as executed`
              );
              return args.in.ids;
            }
          );
        } catch (e) {
          throw new Error(`failed to lock TypeInstances: ${e.message}`);
        } finally {
          neo4jSession.close();
        }
      },
      unlockTypeInstances: async (
        _obj,
        args: LockingTypeInstanceInput,
        context
      ) => {
        const neo4jSession = context.driver.session();
        try {
          return await neo4jSession.writeTransaction(
            async (tx: Transaction) => {
              await switchLocking(
                tx,
                args,
                `
                  MATCH (ti:TypeInstance)
                  WHERE ti.id IN $in.ids
                  SET ti.lockedBy = null
                  RETURN true as executed`
              );
              return args.in.ids;
            }
          );
        } catch (e) {
          throw new Error(`failed to unlock TypeInstances: ${e.message}`);
        } finally {
          neo4jSession.close();
        }
      },
    },
  },
  config: {
    query: false,
    mutation: false,
  },
});

async function switchLocking(
  tx: Transaction,
  args: LockingTypeInstanceInput,
  executeQuery: string
) {
  const instanceLockedByOthers = await tx.run(
    `MATCH (ti:TypeInstance)
          WHERE ti.id IN $in.ids 
          WITH collect(ti) as allIDs
          
          // Check if all TypeInstances were found
          CALL apoc.when(
              size(allIDs) < size($in.ids),
              'RETURN true as notFoundErr',
              'RETURN false as notFoundErr',
              {in: $in, allIDs: allIDs}
          ) YIELD value as checkIDs
          
          // Check if given TypeInstances are not already locked by others
          CALL {
              MATCH (ti:TypeInstance)
              WHERE ti.id IN $in.ids AND ti.lockedBy IS NOT NULL AND ti.lockedBy <> $in.ownerID
              WITH collect(ti) as lockedIDs
              RETURN lockedIDs
          }
          
          // Execute lock only if all TypeInstance were found and none of them are already locked by another owner
          WITH *
          CALL apoc.do.when(
              size(lockedIDs) > 0 OR checkIDs.notFoundErr,
              '
                  RETURN false as executed
              ',
              '
                  ${executeQuery}
              ',
              {in: $in, checkIDs: checkIDs, lockedIDs: lockedIDs}
          ) YIELD value as lockingProcess
          
          RETURN  allIDs, lockedIDs, lockingProcess`,
    { in: args.in }
  );

  if (!instanceLockedByOthers.records.length) {
    throw new Error(`Internal Server Error, result row is undefined`);
  }

  const record = instanceLockedByOthers.records[0];

  const resultRow: LockingResult = {
    allIDs: record.get("allIDs"),
    lockedIDs: record.get("lockedIDs"),
    lockingProcess: record.get("lockingProcess"),
  };

  validateLockingProcess(resultRow, args.in.ids);
}

function validateLockingProcess(result: LockingResult, expIDs: [string]) {
  if (!result.lockingProcess.executed) {
    const errMsg: string[] = [];

    const foundIDs = result.allIDs.map((item) => item.properties.id);
    const notFoundIDs = expIDs.filter((x) => !foundIDs.includes(x));
    if (notFoundIDs.length !== 0) {
      errMsg.push(
        `TypeInstances with IDs "${notFoundIDs.join('", "')}" were not found`
      );
    }

    const lockedIDs = result.lockedIDs.map((item) => item.properties.id);
    if (lockedIDs.length !== 0) {
      errMsg.push(
        `TypeInstances with IDs "${lockedIDs.join(
          '", "'
        )}" are locked by different owner`
      );
    }

    switch (errMsg.length) {
      case 0:
        break;
      case 1:
        throw new Error(`1 error occurred: ${errMsg.join(", ")}`);
      default:
        throw new Error(
          `${errMsg.length} errors occurred: [${errMsg.join(", ")}]`
        );
    }
  }
}

// In Cypher we throw custom errors, e.g.:
// 	CALL apoc.util.validate(size(foundIDs) < size(allInputIDs), apoc.convert.toJson({code: 404, foundIDs: foundIDs}), null)
//
// which produce such output:
//  Failed to invoke procedure `apoc.cypher.doIt`: Caused by: java.lang.RuntimeException: {"lockedIDs":["b0283e96-ce83-451c-9325-0d144b9cea6a"],"code":409}
//
// This function tries to extract this error, if not possible, returns `null`.
function tryToExtractCustomError(
  gotErr: Error
): UpdateTypeInstanceError | null {
  const firstOpen = gotErr.message.indexOf("{");
  const firstClose = gotErr.message.lastIndexOf("}");
  const candidate = gotErr.message.substring(firstOpen, firstClose + 1);

  try {
    return JSON.parse(candidate);
  } catch (e) {
    /* cannot extract, return generic error */
  }

  return null;
}
