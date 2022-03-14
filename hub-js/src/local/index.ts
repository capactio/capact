import { readFileSync } from "fs";
import { makeAugmentedSchema } from "neo4j-graphql-js";
import { createTypeInstances } from "./resolver/mutation/create-type-instances";
import { updateTypeInstances } from "./resolver/mutation/update-type-instances";
import { deleteTypeInstance } from "./resolver/mutation/delete-type-instance";
import { createTypeInstance } from "./resolver/mutation/create-type-instance";
import { lockTypeInstances } from "./resolver/mutation/lock-type-instances";
import { unlockTypeInstances } from "./resolver/mutation/unlock-type-instances";
import { typeInstanceResourceVersionSpecValueField } from "./resolver/field/spec-value-field";

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
    TypeInstanceResourceVersionSpec: {
      value: typeInstanceResourceVersionSpecValueField,
    },
  },
  config: {
    query: false,
    mutation: false,
  },
});
