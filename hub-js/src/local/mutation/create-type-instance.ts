import { Context } from "./context";
import {
  CreateTypeInstanceInput,
  CreateTypeInstancesInput,
  TypeInstanceUsesRelationInput,
} from "../types/type-instance";
import { createTypeInstances } from "./create-type-instances";

interface createTypeInstancesArgs {
  in: CreateTypeInstanceInput;
}

export async function createTypeInstance(
  obj: any,
  args: createTypeInstancesArgs,
  context: Context
) {
  const input = {
    in: {
      typeInstances: [args.in] as CreateTypeInstanceInput[],
      usesRelations: [] as TypeInstanceUsesRelationInput[],
    } as CreateTypeInstancesInput,
  };
  try {
    const result = await createTypeInstances(obj, input, context);
    return result[0];
  } catch (e) {
    const err = e as Error;
    throw new Error(`failed to create TypeInstance: ${err.message}`);
  }
}
