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
      updateTypeInstance: async (_obj) => {
        return Object({
          id: "4123-mocked-id",
          typeRef: {
            path: "cap.mocked.update.type.instance",
            revision: "0.1.1",
          },
          latestResourceVersion: {
            resourceVersion: 1,
            metadata: {
              id: "dba53e7e-2249-4d10-854f-4803b331313f",
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
              id: "dba53e7e-2249-4d10-854f-4803b331313f",
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
                id: "dba53e7e-2249-4d10-854f-4803b331313f",
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
              id: "dba53e7e-2249-4d10-854f-4803b331313f",
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
        });
      },
    },
  },
  config: {
    query: false,
    mutation: false,
  },
});
