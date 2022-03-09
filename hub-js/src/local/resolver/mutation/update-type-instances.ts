import { cypherMutation } from "neo4j-graphql-js";
import { GraphQLResolveInfo } from "graphql";
import _ from "lodash";
import neo4j, { QueryResult, Transaction } from "neo4j-driver";
import {
  CustomCypherErrorCode,
  CustomCypherErrorOutput,
  tryToExtractCustomCypherError,
} from "./cypher-errors";
import { logger } from "../../../logger";
import { Context } from "./context";
import { Operation } from "../../storage/update-args-context";

interface UpdateTypeInstancesInput {
  in: [
    {
      id: string;
      typeInstance: {
        value?: undefined;
      };
    }
  ];
}

export async function updateTypeInstances(
  _: undefined,
  args: UpdateTypeInstancesInput,
  context: Context,
  resolveInfo: GraphQLResolveInfo
) {
  logger.debug("Executing query to update TypeInstance(s)", args);

  context.updateArgs.SetOperation(Operation.UpdateTypeInstancesMutation);
  args.in.forEach((x) => {
    context.updateArgs.SetValue(x.id, x.typeInstance.value);
  });

  const neo4jSession = context.driver.session();

  try {
    return await neo4jSession.writeTransaction(async (tx: Transaction) => {
      // NOTE: we need to record for each input TypeInstance's id, current latest
      // revision in order to know for which revision the value property is already known and
      // stored.
      const instancesResult = await getLatestRevisionVersions(tx, args);
      instancesResult.records.forEach((record) => {
        context.updateArgs.SetLastKnownRev(record.get("id"), record.get("ver"));
      });

      const [query, queryParams] = cypherMutation(args, context, resolveInfo);
      const outputResult = await tx.run(query, queryParams);

      return extractUpdateMutationResult(outputResult);
    });
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

// Simplified version of: https://github.com/neo4j-graphql/neo4j-graphql-js/blob/381ef0302bbd11ecd0f94f978045cdbc61c39b8e/src/utils.js#L57
// We know the variable name as the mutation is written by us, and this function is not meant to be generic.
function extractUpdateMutationResult(result: QueryResult) {
  const data = result.records.map((record) => record.get("typeInstance"));
  // handle Integer fields
  return _.cloneDeepWith(data, (field) => {
    if (neo4j.isInt(field)) {
      // See: https://neo4j.com/docs/api/javascript-driver/current/class/src/v1/integer.js~Integer.html
      return field.inSafeRange() ? field.toNumber() : field.toString();
    }
    return;
  });
}

async function getLatestRevisionVersions(
  tx: Transaction,
  args: UpdateTypeInstancesInput
) {
  const typeInstanceIds = args.in.map((x) => x.id);
  return tx.run(
    `
           UNWIND $ids as id
           MATCH (ti:TypeInstance {id: id})

           WITH *
           // Get Latest Revision
           CALL {
               WITH ti
               WITH ti
               MATCH (ti)-[:CONTAINS]->(tir:TypeInstanceResourceVersion)
               RETURN tir ORDER BY tir.resourceVersion DESC LIMIT 1
           }

           RETURN id, tir.resourceVersion as ver
        `,
    { ids: typeInstanceIds }
  );
}
