import { readFileSync } from "fs";
import { makeAugmentedSchema, neo4jgraphql } from "neo4j-graphql-js";
import { Driver, Transaction } from "neo4j-driver";

const typeDefs = readFileSync("./graphql/local/schema.graphql", "utf-8");

interface CreateTypeInstancesArgs {
  in: {
    typeInstances: Array<{ alias: string }>;
    usesRelations: Array<{ from: string; to: string }>;
  };
}

interface ContextWithDriver {
  driver: Driver;
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
               MERGE (typeRef:TypeInstanceTypeReference {path: typeInstance.typeRef.path, revision: typeInstance.typeRef.revision})
               
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
      updateTypeInstances: async (
        obj,
        args: { in: [{ id: string }] },
        context,
        resolveInfo
      ) => {
        const neo4jSession = context.driver.session();
        try {
          const ids = args.in.map((item) => item.id);

          return await neo4jSession.writeTransaction(
            async (tx: Transaction) => {
              const updateTypeInstancesResult = await tx.run(
                `
                    OPTIONAL MATCH (ti:TypeInstance)
                    WHERE ti.id IN $ids 
                    RETURN ti.id as foundIds`,
                { ids }
              );

              const extractedResult = updateTypeInstancesResult.records.map(
                (record) => record.get("foundIds")
              );
              const notFoundIDs = ids.filter(
                (x) => !extractedResult.includes(x)
              );

              if (notFoundIDs.length !== 0) {
                throw new Error(
                  `TypeInstance with ID(s) "${notFoundIDs.join(
                    ", "
                  )}" not found`
                );
              }
              return neo4jgraphql(obj, args, context, resolveInfo);
            }
          );
        } catch (e) {
          throw new Error(`failed to update TypeInstances": ${e.message}`);
        } finally {
          neo4jSession.close();
        }
      },
      updateTypeInstance: async (obj, args, context, resolveInfo) => {
        const data = await neo4jgraphql(obj, args, context, resolveInfo);
        if (data === null) {
          return new Error(
            `failed to update TypeInstance with ID "${args.id}": TypeInstance not found`
          );
        }
        return data;
      },
      deleteTypeInstance: async (_obj, args, context) => {
        const neo4jSession = context.driver.session();
        try {
          return await neo4jSession.writeTransaction(
            async (tx: Transaction) => {
              const deleteTypeInstanceResult = await tx.run(
                `
                    MATCH (ti:TypeInstance {id: $id})-[:CONTAINS]->(tirs: TypeInstanceResourceVersion)
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
