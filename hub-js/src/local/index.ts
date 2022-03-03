import { readFileSync } from "fs";
import { makeAugmentedSchema } from "neo4j-graphql-js";
import { createTypeInstances } from "./mutation/create-type-instances";
import { updateTypeInstances } from "./mutation/update-type-instances";
import { deleteTypeInstance } from "./mutation/delete-type-instance";
import { createTypeInstance } from "./mutation/create-type-instance";
import { lockTypeInstances } from "./mutation/lock-type-instances";
import { unlockTypeInstances } from "./mutation/unlock-type-instances";

const typeDefs = readFileSync("./graphql/local/schema.graphql", "utf-8");

export const schema = makeAugmentedSchema({
  typeDefs,
  resolvers: {
    Mutation: {
      createTypeInstances,
      createTypeInstance,
      updateTypeInstances,
      deleteTypeInstance,
      lockTypeInstances,
      unlockTypeInstances,
    },
  },
  config: {
    query: false,
    mutation: false,
  },
});
