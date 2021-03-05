import { GraphQLSchema } from "graphql";
import { Driver } from "neo4j-driver";
import { assertSchema } from "neo4j-graphql-js";
import { schema as publicSchema } from "./public";
import { schema as localSchema } from "./local";
import { schema as localV2Schema } from "./local-v2";

export enum OCHMode {
  Local = "local",
  Public = "public",
  LocalV2 = "local-v2",
}

export function getSchemaForMode(mode: string): GraphQLSchema {
  switch (mode) {
    case OCHMode.Local:
      return localSchema;

    case OCHMode.LocalV2:
      return localV2Schema;

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
