import { readFileSync } from 'fs';
import { makeAugmentedSchema } from 'neo4j-graphql-js';

const typeDefs = readFileSync('./graphql/public.graphql', 'utf-8');

const schema = makeAugmentedSchema({
  typeDefs,
  config: {
    query: false,
    mutation: false,
  },
});

export default schema;
