import { readFileSync } from "fs";
import { makeAugmentedSchema } from "neo4j-graphql-js";
import { Driver, Transaction } from "neo4j-driver";

const typeDefs = readFileSync("./graphql/local-v2/schema.graphql", "utf-8");

interface UpdateTypeInstancesInput {
  in: [{ id: string }];
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

function fixTypeInstance(id: string) {
  return {
    id,
    typeRef: {
      path: "cap.mocked.update.type.instance",
      revision: "0.1.1",
    },
    latestResourceVersion: {
      resourceVersion: 1,
      metadata: {
        id,
        attributes: [
          {
            path: "cap.attribute.sample",
            revision: "0.1.0",
          },
        ],
      },
      spec: {
        typeRef: {
          path: "cap.type.sample",
          revision: "0.1.0",
        },
        value: {
          hello: "world",
        },
      },
    },
    firstResourceVersion: {
      resourceVersion: 1,
      metadata: {
        id,
        attributes: [
          {
            path: "cap.attribute.sample",
            revision: "0.1.0",
          },
        ],
      },
      spec: {
        typeRef: {
          path: "cap.type.sample",
          revision: "0.1.0",
        },
        value: {
          hello: "world",
        },
      },
    },
    previousResourceVersion: null,
    resourceVersions: [
      {
        resourceVersion: 1,
        metadata: {
          id,
          attributes: [
            {
              path: "cap.attribute.sample",
              revision: "0.1.0",
            },
          ],
        },
        spec: {
          typeRef: {
            path: "cap.type.sample",
            revision: "0.1.0",
          },
          value: {
            hello: "world",
          },
        },
      },
    ],
    resourceVersion: {
      resourceVersion: 1,
      metadata: {
        id,
        attributes: [
          {
            path: "cap.attribute.sample",
            revision: "0.1.0",
          },
        ],
      },
      spec: {
        typeRef: {
          path: "cap.type.sample",
          revision: "0.1.0",
        },
        value: {
          hello: "world",
        },
      },
    },
  };
}

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
               MERGE (typeRef:TypeReference {path: typeInstance.typeRef.path, revision: typeInstance.typeRef.revision})
               
               CREATE (ti:TypeInstance {id: apoc.create.uuid()})
               CREATE (ti)-[:OF_TYPE]->(typeRef)
               
               CREATE (tir: TypeInstanceResourceVersion {resourceVersion: 1})
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
      updateTypeInstance: async (_obj, args) => fixTypeInstance(args.id),
      updateTypeInstances: async (_obj, args: UpdateTypeInstancesInput) => {
        const ti = args.in.map((elem) => fixTypeInstance(elem.id));
        return ti;
      },
      deleteTypeInstance: async (_obj, args, context) => {
        const neo4jSession = context.driver.session();
        try {
          return await neo4jSession.writeTransaction(
            async (tx: Transaction) => {
              const deleteTypeInstanceResult = await tx.run(
                `
                    MATCH (ti:TypeInstance {id: $id})-[:CONTAINS]->(tirs: TypeInstanceResourceVersion)
                    MATCH (ti)-[:OF_TYPE]->(typeRef: TypeReference)
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
                { id: args.id }
              );

              if (
                !deleteTypeInstanceResult.summary.counters.containsUpdates()
              ) {
                throw new Error(`TypeInstance not found`);
              }
              return args.id;
            }
          );
        } catch (e) {
          throw new Error(
            `failed to delete TypeInstance with ID "${args.id}": ${e.message}`
          );
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
