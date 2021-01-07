import { readFileSync } from 'fs';
import { makeAugmentedSchema } from 'neo4j-graphql-js';

const typeDefs = readFileSync('./graphql/local.graphql', 'utf-8');

const schema = makeAugmentedSchema({
  typeDefs,
  config: {
    query: true,
    mutation: true,
  },
});

export default schema;
