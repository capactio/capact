import { readFileSync } from 'fs';
import { makeAugmentedSchema } from 'neo4j-graphql-js';
import { IResolvers } from 'graphql-tools';

const typeDefs = readFileSync('./graphql/public.graphql', 'utf-8');

const nameResolver = (object: { path: string }) => object.path.split('.').pop();

const prefixResolver = (object: { path: string }) => {
  const parts = object.path.split('.');
  return parts.slice(0, parts.length - 1).join('.');
};

const resolvers: IResolvers = {
  RepoMetadata: {
    name: nameResolver,
    prefix: prefixResolver,
  },
  Interface: {
    name: nameResolver,
    prefix: prefixResolver,
  },
  Type: {
    name: nameResolver,
    prefix: prefixResolver,
  },
  Implementation: {
    name: nameResolver,
    prefix: prefixResolver,
  },
  Tag: {
    name: nameResolver,
    prefix: prefixResolver,
  },
  GenericMetadata: {
    name: nameResolver,
    prefix: prefixResolver,
  },
  ImplementationMetadata: {
    name: nameResolver,
    prefix: prefixResolver,
  },
  TypeMetadata: {
    name: nameResolver,
    prefix: prefixResolver,
  },
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
