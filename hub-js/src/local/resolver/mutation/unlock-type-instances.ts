import { Transaction } from "neo4j-driver";
import { Context } from "./context";
import {
  getTypeInstanceStoredExternally,
  LockingTypeInstanceInput,
  switchLocking,
} from "./lock-type-instances";
import { logger } from "../../../logger";

interface UnLockTypeInstanceInput extends LockingTypeInstanceInput {}

export async function unlockTypeInstances(
  _: undefined,
  args: UnLockTypeInstanceInput,
  context: Context
) {
  const neo4jSession = context.driver.session();
  try {
    return await neo4jSession.writeTransaction(async (tx: Transaction) => {
      logger.debug("Executing query to unlock TypeInstance(s)", args);
      await switchLocking(
        tx,
        args,
        `
            MATCH (ti:TypeInstance)
            WHERE ti.id IN $in.ids
            SET ti.lockedBy = null
            RETURN true as executed`
      );
      const unlockExternals = await getTypeInstanceStoredExternally(
        tx,
        args.in.ids,
        args.in.ownerID
      );
      await context.delegatedStorage.Unlock(...unlockExternals);

      return args.in.ids;
    });
  } catch (e) {
    const err = e as Error;
    throw new Error(`failed to unlock TypeInstances: ${err.message}`);
  } finally {
    await neo4jSession.close();
  }
}
