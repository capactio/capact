import { GraphQLSchema } from "graphql";
import { Driver } from "neo4j-driver";
import { assertSchema } from "neo4j-graphql-js";
import { schema as publicSchema } from "./public/public";
import { schema as localSchema } from "./local/local";

export enum HubMode {
  Local = "local",
  Public = "public",
}

export function getSchemaForMode(mode: string): GraphQLSchema {
  switch (mode) {
    case HubMode.Local:
      return localSchema;

    case HubMode.Public:
      return publicSchema;

    default:
      throw Error(`unknown Hub mode: ${mode}`);
  }
}

export const assertSchemaOnDatabase = (schema: GraphQLSchema, driver: Driver) =>
  assertSchema({
    schema,
    driver,
  });
