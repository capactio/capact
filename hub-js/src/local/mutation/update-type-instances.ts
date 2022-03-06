import { Neo4jContext, neo4jgraphql } from "neo4j-graphql-js";
import { GraphQLResolveInfo } from "graphql";
import {
  CustomCypherErrorCode,
  CustomCypherErrorOutput,
  tryToExtractCustomCypherError,
} from "./cypher-errors";
import { logger } from "../../logger";

export async function updateTypeInstances(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  obj: any,
  args: { in: [{ id: string }] },
  context: Neo4jContext,
  resolveInfo: GraphQLResolveInfo
) {
  try {
    logger.debug("Executing query to update TypeInstance(s)", args);
    return await neo4jgraphql(obj, args, context, resolveInfo);
  } catch (e) {
    let err = e as Error;

    const customErr = tryToExtractCustomCypherError(err);
    if (customErr) {
      switch (customErr.code) {
        case CustomCypherErrorCode.Conflict:
          err = generateConflictError(customErr);
          break;
        case CustomCypherErrorCode.NotFound: {
          err = generateNotFoundError(args.in, customErr);
          break;
        }
        default:
          err = Error(`Unexpected error code ${customErr.code}`);
          break;
      }
    }
    throw new Error(`failed to update TypeInstances: ${err.message}`);
  }
}

function generateNotFoundError(
  input: [{ id: string }],
  customErr: CustomCypherErrorOutput
) {
  const ids = input.map(({ id }) => id);
  const notFoundIDs = ids
    .filter((x) => !customErr.ids.includes(x))
    .join(`", "`);
  return Error(`TypeInstances with IDs "${notFoundIDs}" were not found`);
}

function generateConflictError(customErr: CustomCypherErrorOutput) {
  if (!Object.prototype.hasOwnProperty.call(customErr, "ids")) {
    // it shouldn't happen
    return Error(`TypeInstances are locked by different owner`);
  }
  const conflictIDs = customErr.ids.join(`", "`);
  return Error(
    `TypeInstances with IDs "${conflictIDs}" are locked by different owner`
  );
}
