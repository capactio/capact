import { readFileSync } from "fs";
import { makeAugmentedSchema } from "neo4j-graphql-js";
import { Transaction } from "neo4j-driver";

const typeDefs = readFileSync("./graphql/local/schema.graphql", "utf-8");

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
          return await neo4jSession.writeTransaction(
            async (tx: Transaction) => {
              const createTypeInstanceResult = await tx.run(
                `UNWIND $typeInstances AS typeInstance
               CREATE (ti: TypeInstance {resourceVersion: 1})
               CREATE (ti)-[:DESCRIBED_BY]->(metadata: TypeInstanceMetadata {id: apoc.create.uuid()})
               CREATE (ti)-[:SPECIFIED_BY]->(spec: TypeInstanceSpec {value: apoc.convert.toJson(typeInstance.value)})
               CREATE (spec)-[:OF_TYPE]->(typeRef: TypeReference {path: typeInstance.typeRef.path, revision: typeInstance.typeRef.revision})

               FOREACH (attr in typeInstance.attributes |
                 CREATE (metadata)-[:CHARACTERIZED_BY]->(attrRef: AttributeReference {path: attr.path, revision: attr.revision})
               )

               RETURN metadata.id as uuid, typeInstance.alias as alias
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
                (acc: { [key: string]: string }, cur) => ({
                  ...acc,
                  [cur.get("alias") || cur.get("uuid")]: cur.get("uuid"),
                }),
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
               MATCH (fromTi:TypeInstance)-[:DESCRIBED_BY]->(:TypeInstanceMetadata {id: usesRelation.from})
               MATCH (toTi:TypeInstance)-[:DESCRIBED_BY]->(:TypeInstanceMetadata {id: usesRelation.to})
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
