import { readFileSync } from "fs";
import { makeAugmentedSchema } from "neo4j-graphql-js";
import { logger } from "../logger";

const typeDefs = readFileSync("./graphql/local/schema.graphql", "utf-8");

export const schema = makeAugmentedSchema({
  typeDefs,
  config: {
    query: false,
    mutation: false,
  },
});
