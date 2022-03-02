import { readFileSync } from "fs";
import { makeAugmentedSchema } from "neo4j-graphql-js";
import { createTypeInstances } from "./mutation/create-type-instances";
import { updateTypeInstances } from "./mutation/update-type-instances";
import { deleteTypeInstance } from "./mutation/delete-type-instance";
import { createTypeInstance } from "./mutation/create-type-instance";
import {
  toggleLockTypeInstances,
  unlockTypeInstances
} from "./mutation/toggle-lock-type-instances";

const typeDefs = readFileSync("./graphql/local/schema.graphql", "utf-8");

export const schema = makeAugmentedSchema({
  typeDefs,
  resolvers: {
    Mutation: {
      createTypeInstances: createTypeInstances,
      createTypeInstance: createTypeInstance,
      updateTypeInstances: updateTypeInstances,
      deleteTypeInstance: deleteTypeInstance,
      lockTypeInstances: toggleLockTypeInstances,
      unlockTypeInstances: unlockTypeInstances
    }
  },
  config: {
    query: false,
    mutation: false
  }
});
