import { Neo4jContext, neo4jgraphql } from "neo4j-graphql-js";
import { GraphQLResolveInfo } from "graphql";
import { tryToExtractCustomError } from "./helpers";

export enum UpdateTypeInstanceErrorCode {
  Conflict = 409,
  NotFound = 404,
}

export interface UpdateTypeInstanceError {
  code: UpdateTypeInstanceErrorCode;
  ids: string[];
}

export async function updateTypeInstances(
  obj: any,
  args: { in: [{ id: string }] },
  context: Neo4jContext,
  resolveInfo: GraphQLResolveInfo
) {
  try {
    return await neo4jgraphql(obj, args, context, resolveInfo);
  } catch (e) {
    let err = e as Error;

    const customErr = tryToExtractCustomError(err);
    if (customErr) {
      switch (customErr.code) {
        case UpdateTypeInstanceErrorCode.Conflict:
          const conflictIDs = customErr.ids.join(`", "`);
          err = Error(
            `TypeInstances with IDs "${conflictIDs}" are locked by different owner`
          );
          break;
        case UpdateTypeInstanceErrorCode.NotFound: {
          const ids = args.in.map(({ id }) => id);
          const notFoundIDs = ids
            .filter((x) => !customErr.ids.includes(x))
            .join(`", "`);
          err = Error(`TypeInstances with IDs "${notFoundIDs}" were not found`);
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
