import { GraphQLSchema } from "graphql";
import { Driver } from "neo4j-driver";
import { assertSchema } from "neo4j-graphql-js";
import { schema as publicSchema } from "./public";
import { schema as localSchema } from "./local";

enum OCHMode {
  Local = "local",
  Public = "public",
}

export function getSchemaForMode(mode: string): GraphQLSchema {
  switch (mode) {
    case OCHMode.Local:
      return localSchema;

    case OCHMode.Public:
      return publicSchema;

    default:
      throw Error(`unknown OCH mode: ${mode}`);
  }
}

export const assertSchemaOnDatabase = (schema: GraphQLSchema, driver: Driver) =>
  assertSchema({
    schema,
    driver,
  });
