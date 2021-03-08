import { readFileSync } from "fs";
import { makeAugmentedSchema } from "neo4j-graphql-js";
import { Transaction } from "neo4j-driver";

const typeDefs = readFileSync("./graphql/local-v2/schema.graphql", "utf-8");

interface CreateTypeInstancesArgs {
  in: {
    typeInstances: Array<{ alias: string }>;
    usesRelations: Array<{ from: string; to: string }>;
  };
}

interface UpdateTypeInstancesInput {
  in: [{ id: string }];
}

function fixTypeInstance(id: string) {
  return {
    id: id,
    typeRef: {
      path: "cap.mocked.update.type.instance",
      revision: "0.1.1",
    },
    latestResourceVersion: {
      resourceVersion: 1,
      metadata: {
        id: id,
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
        id: id,
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
          id: id,
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
        id: id,
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
        _obj,
        args: CreateTypeInstancesArgs,
        context
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
          return await neo4jSession.writeTransaction(async (_: Transaction) => {
            return aliases.map((entry) =>
              Object({
                alias: entry,
                id: `4123-mocked-id-for-${entry}`,
              })
            );
          });
        } catch (e) {
          throw new Error(`failed to create the TypeInstances: ${e.message}`);
        } finally {
          neo4jSession.close();
        }
      },
      updateTypeInstance: async (_obj, args) => {
        return Object(fixTypeInstance(args.id));
      },
      updateTypeInstances: async (_obj, args: UpdateTypeInstancesInput) => {
        const ti = args.in.map((elem) => fixTypeInstance(elem.id));
        return Object(ti);
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
              // Always return ID even if not found, request should be idempotent

              if (deleteTypeInstanceResult.records.length === 0) {
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
