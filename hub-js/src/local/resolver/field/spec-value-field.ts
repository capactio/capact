import { logger } from "../../../logger";
import { GetInput } from "../../storage/service";
import { Context } from "../mutation/context";

// Represents contract defined on `TypeInstanceResourceVersionSpec.Value` field cypher query.
interface InputObject {
  value: {
    // specifies whether data is stored in built-in or external storage
    abstract: boolean;
    // holds the TypeInstance's value stored in built-in storage
    builtinValue: undefined;
    // holds information needed to fetch the TypeInstance's value from external storage
    fetchInput: GetInput;
  };
}

export async function typeInstanceResourceVersionSpecValueField(
  { value: obj }: InputObject,
  _: undefined,
  context: Context
) {
  logger.debug("Executing custom field resolver for 'value' field", obj);
  if (obj.abstract) {
    logger.debug("Return data stored in built-in storage");
    return obj.builtinValue;
  } else {
    logger.debug("Return data stored in external storage");
    const resp = await context.delegatedStorage.Get(obj.fetchInput);
    return resp[obj.fetchInput.typeInstance.id];
  }
}
