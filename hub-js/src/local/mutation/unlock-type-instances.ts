import { Transaction } from "neo4j-driver";
import { ContextWithDriver } from "./context";
import { LockingTypeInstanceInput, switchLocking } from "./lock-type-instances";
import { logger } from "../../logger";

interface UnLockTypeInstanceInput extends LockingTypeInstanceInput {}

export async function unlockTypeInstances(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  _: any,
  args: UnLockTypeInstanceInput,
  context: ContextWithDriver
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
      return args.in.ids;
    });
  } catch (e) {
    const err = e as Error;
    throw new Error(`failed to unlock TypeInstances: ${err.message}`);
  } finally {
    await neo4jSession.close();
  }
}
