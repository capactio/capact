import { readFileSync } from "fs";
import { makeAugmentedSchema } from "neo4j-graphql-js";

const typeDefs = readFileSync("./graphql/local/schema.graphql", "utf-8");

export const schema = makeAugmentedSchema({
  typeDefs,
  resolvers: {
    Mutation: {
      createTypeInstances: async (_obj, args, context) => {
        const { typeInstances, usesRelations } = args.in;

        let result: any;

        const neo4jSession = context.driver.session();

        try {
          result = await neo4jSession.writeTransaction(async (tx: any) => {
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

            const aliasMappings = createTypeInstanceResult.records
              .map((record: any) =>
                Object({
                  [record.get("alias")]: record.get("uuid"),
                })
              )
              .reduce((a: object, b: object) => Object.assign(a, b));

            const usesRelationsParams = usesRelations.map(({ from, to }: any) =>
              Object({
                from: aliasMappings[from] || from,
                to: aliasMappings[to] || to,
              })
            );

            await tx.run(
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

            return Object.values(aliasMappings);
          });
        } finally {
          neo4jSession.close();
        }

        return result;
      },
    },
  },
  config: {
    query: false,
    mutation: false,
  },
});
