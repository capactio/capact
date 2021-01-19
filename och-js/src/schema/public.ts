import { readFileSync } from 'fs';
import { makeAugmentedSchema, neo4jgraphql } from 'neo4j-graphql-js';
import { IResolvers } from 'graphql-tools';
import { GraphQLResolveInfo } from 'graphql/type';

const typeDefs = readFileSync('./graphql/public.graphql', 'utf-8');

const nameResolver = (object: { path: string }) => object.path.split('.').pop();

const prefixResolver = (object: { path: string }, params, context, resolveInfo) => {
  const parts = object.path.split('.');
  return parts.slice(0, parts.length - 1).join('.');
};

const resolvers: IResolvers = {
};

const schema = makeAugmentedSchema({
  typeDefs,
  resolvers,
  config: {
    query: false,
    mutation: false,
  },
});

export default schema;
