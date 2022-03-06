import { Transaction } from "neo4j-driver";
import { ContextWithDriver } from "./context";
import { logger } from "../../logger";

export interface LockingTypeInstanceInput {
  in: {
    ids: [string];
    ownerID: string;
  };
}

interface TypeInstanceNode {
  properties: { id: string; lockedBy: string };
}

interface LockingResult {
  allIDs: [TypeInstanceNode];
  lockedIDs: [TypeInstanceNode];
  lockingProcess: {
    executed: boolean;
  };
}

export async function lockTypeInstances(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  _: any,
  args: LockingTypeInstanceInput,
  context: ContextWithDriver
) {
  const neo4jSession = context.driver.session();
  try {
    return await neo4jSession.writeTransaction(async (tx: Transaction) => {
      logger.debug("Executing query to lock TypeInstance(s)", args);
      await switchLocking(
        tx,
        args,
        `
            MATCH (ti:TypeInstance)
            WHERE ti.id IN $in.ids
            SET ti.lockedBy = $in.ownerID
            RETURN true as executed`
      );
      return args.in.ids;
    });
  } catch (e) {
    const err = e as Error;
    throw new Error(`failed to lock TypeInstances: ${err.message}`);
  } finally {
    await neo4jSession.close();
  }
}

export async function switchLocking(
  tx: Transaction,
  args: LockingTypeInstanceInput,
  executeQuery: string
) {
  const instanceLockedByOthers = await tx.run(
    `MATCH (ti:TypeInstance)
          WHERE ti.id IN $in.ids
          WITH collect(ti) as allIDs

          // Check if all TypeInstances were found
          CALL apoc.when(
              size(allIDs) < size($in.ids),
              'RETURN true as notFoundErr',
              'RETURN false as notFoundErr',
              {in: $in, allIDs: allIDs}
          ) YIELD value as checkIDs

          // Check if given TypeInstances are not already locked by others
          CALL {
              MATCH (ti:TypeInstance)
              WHERE ti.id IN $in.ids AND ti.lockedBy IS NOT NULL AND ti.lockedBy <> $in.ownerID
              WITH collect(ti) as lockedIDs
              RETURN lockedIDs
          }

          // Execute lock only if all TypeInstance were found and none of them are already locked by another owner
          WITH *
          CALL apoc.do.when(
              size(lockedIDs) > 0 OR checkIDs.notFoundErr,
              '
                  RETURN false as executed
              ',
              '
                  ${executeQuery}
              ',
              {in: $in, checkIDs: checkIDs, lockedIDs: lockedIDs}
          ) YIELD value as lockingProcess

          RETURN  allIDs, lockedIDs, lockingProcess`,
    { in: args.in }
  );

  if (!instanceLockedByOthers.records.length) {
    throw new Error(`Internal Server Error, result row is undefined`);
  }

  const record = instanceLockedByOthers.records[0];

  const resultRow: LockingResult = {
    allIDs: record.get("allIDs"),
    lockedIDs: record.get("lockedIDs"),
    lockingProcess: record.get("lockingProcess"),
  };

  validateLockingProcess(resultRow, args.in.ids);
}

function validateLockingProcess(result: LockingResult, expIDs: [string]) {
  if (!result.lockingProcess.executed) {
    const errMsg: string[] = [];

    const foundIDs = result.allIDs.map((item) => item.properties.id);
    const notFoundIDs = expIDs.filter((x) => !foundIDs.includes(x));
    if (notFoundIDs.length !== 0) {
      errMsg.push(
        `TypeInstances with IDs "${notFoundIDs.join('", "')}" were not found`
      );
    }

    const lockedIDs = result.lockedIDs.map((item) => item.properties.id);
    if (lockedIDs.length !== 0) {
      errMsg.push(
        `TypeInstances with IDs "${lockedIDs.join(
          '", "'
        )}" are locked by different owner`
      );
    }

    switch (errMsg.length) {
      case 0:
        break;
      case 1:
        throw new Error(`1 error occurred: ${errMsg.join(", ")}`);
      default:
        throw new Error(
          `${errMsg.length} errors occurred: [${errMsg.join(", ")}]`
        );
    }
  }
}
