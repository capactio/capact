import { logger } from "../../../logger";
import { GetInput } from "../../storage/service";
import { Context } from "../mutation/context";
import { Operation } from "../../storage/update-args-container";
import _ from "lodash";
import { Mutex } from "async-mutex";

const mutex = new Mutex();

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
  // This is a field resolver, it can be called multiple times within the same `query/mutation`.
  // We also perform, external calls to change the state if needed, due to that fact, we use
  // mutex to ensure that we won't call backend multiple times as backends may be not thread safe.
  return await mutex.runExclusive(async () => {
    logger.debug("Executing custom field resolver for 'value' field", obj);
    if (obj.abstract) {
      logger.debug("Return data stored in built-in storage");
      return obj.builtinValue;
    }

    switch (context.updateArgs.GetOperation()) {
      case Operation.UpdateTypeInstancesMutation:
        return await resolveMutationReturnValue(context, obj.fetchInput);
      default: {
        logger.debug("Return data stored in external storage");
        const resp = await context.delegatedStorage.Get(obj.fetchInput);
        return resp[obj.fetchInput.typeInstance.id];
      }
    }
  });
}

async function resolveMutationReturnValue(
  context: Context,
  fetchInput: GetInput
) {
  const tiId = fetchInput.typeInstance.id;
  const revToResolve = fetchInput.typeInstance.resourceVersion;

  let newValue = context.updateArgs.GetValue(tiId);
  const lastKnownRev = context.updateArgs.GetLastKnownRev(tiId);

  // During the mutation someone asked to return also:
  // - `firstResourceVersion`
  // - and/or `previousResourceVersion`
  // - and/or `resourceVersion` with already known revision
  // - and/or `resourceVersions` which holds also previous already stored revisions
  if (revToResolve <= lastKnownRev) {
    logger.debug(
      "Fetch data from external storage for already known revision",
      fetchInput
    );
    const resp = await context.delegatedStorage.Get(fetchInput);
    return resp[tiId];
  }

  // If the revision is higher that the last known revision version, it means that we need to store that into delegated
  // storage.

  // 1. Based on our contract, if user didn't provide value, we need to fetch the old one and put it
  // to the new revision.
  if (!newValue) {
    const previousValue: GetInput = _.cloneDeep(fetchInput);
    previousValue.typeInstance.resourceVersion -= 1;

    logger.debug(
      "Fetching previous value from external storage",
      previousValue
    );
    const resp = await context.delegatedStorage.Get(previousValue);
    newValue = resp[tiId];
  }

  console.log(context.updateArgs.GetOwnerID(fetchInput.typeInstance.id));
  // 2. Update TypeInstance's value
  const update = {
    backend: fetchInput.backend,
    typeInstance: {
      id: fetchInput.typeInstance.id,
      newResourceVersion: fetchInput.typeInstance.resourceVersion,
      newValue,
      ownerID: context.updateArgs.GetOwnerID(fetchInput.typeInstance.id),
    },
  };

  logger.debug("Storing new value into external storage", update);
  await context.delegatedStorage.Update(update);

  // 3. Update last known revision, so if `value` resolver is called next time we won't update it once again
  // and run into `ALREADY_EXISTS` error.
  context.updateArgs.SetLastKnownRev(
    update.typeInstance.id,
    update.typeInstance.newResourceVersion
  );

  return newValue;
}
