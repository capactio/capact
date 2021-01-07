import { GraphQLSchema } from 'graphql';
import { Driver } from 'neo4j-driver';
import { assertSchema } from 'neo4j-graphql-js';
import publicSchema from './public';
import localSchema from './local';

const LocalMode = 'local';
const PublicMode = 'public';

export function getSchemaForMode(mode: string): GraphQLSchema {
  switch (mode) {
    case LocalMode:
      return localSchema;

    case PublicMode:
      return publicSchema;

    default:
      throw Error(`unknown OCH mode: ${mode}`);
  }
}

export const assertSchemaOnDatabase = (schema: GraphQLSchema, driver: Driver) => assertSchema({
  schema,
  driver,
});

export default {
  getSchemaForMode,
  assertSchemaOnDatabase,
};
