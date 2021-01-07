import { readFileSync } from 'fs';
import { makeAugmentedSchema } from 'neo4j-graphql-js';
import { IResolvers } from 'graphql-tools';

import { runSingleQuery } from '../neo4j';

const typeDefs = readFileSync('./graphql/public.graphql', 'utf-8');

const resolvers: IResolvers = {
  InterfaceRevision: {
    implementationRevisionsForRequirementsCustom:
      (object, params: { [argName: string]: any }, context) => {
      // Transform requirements from array of objects to array of strings
        let typePaths = [];
        if (params && params.filter && params.filter.requirementsSatisfiedBy
        && params.filter.requirementsSatisfiedBy.length > 0) {
          const reqs = params.filter.requirementsSatisfiedBy;
          typePaths = reqs.map(({ typeRef }) => typeRef && typeRef.path);
        }

        const query = `
                MATCH (this:InterfaceRevision)

                // When Implementation doesn't require anything
                CALL{
                WITH this
                MATCH (implRev:ImplementationRevision)-[:IMPLEMENTS]->(this),
                  (implRev)-[:SPECIFIED_BY]->(implRevSpec:ImplementationSpec)
                WHERE NOT (implRevSpec)-[:REQUIRES]->(:ImplementationRequirement)
                RETURN implRev
            
                UNION
            
                // When Implementation has requirements using oneOf
                WITH this
                MATCH (implRev:ImplementationRevision)-[:IMPLEMENTS]->(this),
                  (implRev)-[:SPECIFIED_BY]->(implRevSpec:ImplementationSpec)-[:REQUIRES]->
                  (:ImplementationRequirement)-[:ONE_OF]->(reqItem:ImplementationRequirementItem)
            
                // TODO: hardcoded typeRefPath
                // we could use https://stackoverflow.com/questions/51208263/pass-set-of-parameters-to-neo4j-query
                WHERE reqItem.typeRefPath IN $typePaths
                RETURN implRev
                }

                RETURN implRev
                `;

        return runSingleQuery(context, query, {
          typePaths,
        });
      },
  },
};

const schema = makeAugmentedSchema({
  typeDefs,
  resolvers,
  config: {
    query: true,
    mutation: true,
  },
});

export default schema;
