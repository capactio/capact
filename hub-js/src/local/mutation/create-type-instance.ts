import { Context } from "./context";
import { CreateTypeInstanceInput } from "../types/type-instance";
import {
  createTypeInstances,
  CreateTypeInstancesArgs,
} from "./create-type-instances";

interface CreateTypeInstanceArgs {
  in: CreateTypeInstanceInput;
}

export async function createTypeInstance(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  _: any,
  args: CreateTypeInstanceArgs,
  context: Context
) {
  const input: CreateTypeInstancesArgs = {
    in: {
      typeInstances: [args.in],
      usesRelations: [],
    },
  };
  try {
    const result = await createTypeInstances(_, input, context);
    return result[0].id;
  } catch (e) {
    const err = e as Error;
    throw new Error(`failed to create TypeInstance: ${err.message}`);
  }
}
