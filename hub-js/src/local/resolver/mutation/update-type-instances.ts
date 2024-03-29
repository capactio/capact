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
import DelegatedStorageService, {
  DeleteRevisionInput,
  GetInput,
} from "../../storage/service";
import * as grpc from "@grpc/grpc-js";
import { aggregateError } from "./create-type-instances";

interface UpdateTypeInstancesInput {
  in: [
    {
      id: string;
      ownerID?: string;
      typeInstance: {
        value?: unknown;
        backend?: BackendInput;
      };
    }
  ];
}

interface BackendInput {
  context?: unknown;
}

interface UpdateArgs {
  id: string;
  backend?: BackendInput;
  value?: unknown;
  owner?: string;
}

interface StoredValue {
  rollbackInput: DeleteRevisionInput;
  newContext?: unknown;
}

// Represents contract defined on finding whether TypeInstance's value is stored externally or not.
interface ValueReference {
  // Specifies whether data is stored in built-in or external storage.
  abstract: boolean;
  // Holds information needed to fetch the TypeInstance's value from external storage.
  // Available only if abstract == false.
  fetchInput: GetInput;
}

export async function updateTypeInstances(
  _: unknown,
  args: UpdateTypeInstancesInput,
  context: Context,
  resolveInfo: GraphQLResolveInfo
) {
  logger.debug("Executing query to update TypeInstance(s)", args);

  const neo4jSession = context.driver.session();
  const externallyStored: DeleteRevisionInput[] = [];

  try {
    return await neo4jSession.writeTransaction(async (tx: Transaction) => {
      const externalBackends = await extractInformationAboutValueStore(
        tx,
        args
      );

      for (const [index, item] of args.in.entries()) {
        if (!externalBackends.has(item.id)) {
          continue;
        }

        const out = await storeValueExternally(
          {
            id: item.id,
            value: item.typeInstance.value,
            backend: item.typeInstance.backend,
            owner: item.ownerID,
          },
          context.delegatedStorage,
          externalBackends.get(item.id) as GetInput
        );
        externallyStored.push(out.rollbackInput);

        if (!out.newContext) {
          continue;
        }

        logger.debug("Backend contexts was changed by external backend", {
          id: item.id,
          oldContext: item.typeInstance.backend?.context,
          newContext: out.newContext,
        });
        args.in[index].typeInstance.backend = {
          context: out.newContext,
        };
      }

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

    const rollbackErr = await rollbackExternalStoredRevision(
      context.delegatedStorage,
      externallyStored
    );
    err = aggregateError(err, rollbackErr);
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

async function extractInformationAboutValueStore(
  tx: Transaction,
  args: UpdateTypeInstancesInput
): Promise<Map<string, GetInput>> {
  const typeInstanceIds = args.in.map((x) => x.id);
  const instancesResult = await tx.run(
    `
           UNWIND $ids as id
           MATCH (ti:TypeInstance {id: id})

           WITH *
           // Get Latest Revision
           CALL {
               WITH ti
               WITH ti
               MATCH (ti)-[:CONTAINS]->(rev:TypeInstanceResourceVersion)
               RETURN rev ORDER BY rev.resourceVersion DESC LIMIT 1
           }
           MATCH (rev)-[:SPECIFIED_BY]->(spec:TypeInstanceResourceVersionSpec)
           MATCH (spec)-[:WITH_BACKEND]->(backendCtx)
           MATCH (ti)-[:STORED_IN]->(backendRef)

           WITH *
           CALL apoc.when(
               backendRef.abstract,
               '
                   WITH {
                     abstract: backendRef.abstract
                   } AS value
                   RETURN value
               ',
               '
                   WITH {
                     abstract: backendRef.abstract,
                     fetchInput: {
                        typeInstance: { resourceVersion: rev.resourceVersion, id: ti.id },
                        backend: { context: apoc.convert.fromJsonMap(backendCtx.context), id: backendRef.id}
                     }
                   } AS value
                   RETURN value
               ',
               {spec: spec, rev: rev, ti: ti, backendRef: backendRef, backendCtx: backendCtx}
           ) YIELD value as out

           RETURN id, out.value as ref
        `,
    { ids: typeInstanceIds }
  );

  const externalBackends: Map<string, GetInput> = new Map();
  instancesResult.records.forEach((record) => {
    const id = record.get("id");
    const ref: ValueReference = record.get("ref");

    if (ref.abstract) {
      return;
    }
    externalBackends.set(id, ref.fetchInput);
  });

  return externalBackends;
}

// Ensure that data is deleted in case of not committed transaction
async function rollbackExternalStoredRevision(
  delegatedStorage: DelegatedStorageService,
  externallyStored: DeleteRevisionInput[]
): Promise<Error | undefined> {
  try {
    await delegatedStorage.DeleteRevision(...externallyStored);
  } catch (e) {
    const err = e as grpc.ServiceError;
    if (err.code != grpc.status.NOT_FOUND) {
      return new Error(`rollback externally stored revision: ${err.message}`);
    }
  }
  return;
}

async function storeValueExternally(
  args: UpdateArgs,
  delegatedStorage: DelegatedStorageService,
  fetchInput: GetInput
): Promise<StoredValue> {
  // 1. Based on our contract, if user didn't provide value, we need to fetch the old one and put it
  // to the new revision.
  const requiresInputValue = await delegatedStorage.IsValueAllowedByBackend(
    fetchInput.backend.id
  );
  if (!args.value && requiresInputValue) {
    logger.debug("Fetching previous value from external storage", fetchInput);
    const resp = await delegatedStorage.Get(fetchInput);
    args.value = resp[args.id];
  }

  // 2. If user provided context, override it
  if (args.backend?.context) {
    fetchInput.backend.context = args.backend.context;
  }

  // 3. Update TypeInstance's value
  fetchInput.typeInstance.resourceVersion =
    Number(fetchInput.typeInstance.resourceVersion) + 1;
  const update = {
    backend: fetchInput.backend,
    typeInstance: {
      id: fetchInput.typeInstance.id,
      newResourceVersion: fetchInput.typeInstance.resourceVersion,
      newValue: args.value,
      ownerID: args.owner,
    },
  };

  const out: StoredValue = {
    rollbackInput: {
      backend: update.backend,
      typeInstance: {
        id: update.typeInstance.id,
        resourceVersion: update.typeInstance.newResourceVersion,
        ownerID: update.typeInstance.ownerID,
      },
    },
  };

  logger.debug("Storing new value into external storage", update);
  const updatedContexts = await delegatedStorage.Update(update);
  if (Object.keys(updatedContexts).length) {
    out.newContext = updatedContexts[update.typeInstance.id];
  }

  return out;
}
