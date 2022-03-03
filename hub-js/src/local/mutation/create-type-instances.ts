import { QueryResult, Transaction } from "neo4j-driver";
import { BUILTIN_STORAGE_BACKEND_ID } from "../../config";
import { Context } from "./context";
import {
  uniqueNamesGenerator,
  Config,
  adjectives,
  colors,
  animals,
} from "unique-names-generator";
import { DeleteInput, StoreInput, UpdatedContexts } from "../storage/service";
import {
  CreateTypeInstanceInput,
  CreateTypeInstancesInput,
  TypeInstanceBackendDetails,
  TypeInstanceBackendInput,
  TypeInstanceUsesRelationInput,
} from "../types/type-instance";
import {
  CustomCypherErrorCode,
  CustomCypherErrorOutput,
  tryToExtractCustomCypherError,
} from "./cypher-errors";
import { logger } from "../../logger";
import { builtinStorageBackendDetails } from "./register-built-in-storage";

const genAdjsColorsAndAnimals: Config = {
  dictionaries: [adjectives, colors, animals],
  separator: "_",
  length: 3,
};

export interface CreateTypeInstancesArgs {
  in: CreateTypeInstancesInput;
}

interface AliasMapping {
  [key: string]: string;
}

export type TypeInstanceInput = Omit<CreateTypeInstanceInput, "backend"> & {
  backend: TypeInstanceBackendDetails;
};

export async function createTypeInstances(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  _: any,
  args: CreateTypeInstancesArgs,
  context: Context
) {
  const { typeInstances, usesRelations } = args.in;
  const typeInstancesInput = typeInstances.map(setDefaults);
  try {
    validate(typeInstancesInput);
  } catch (e) {
    const err = e as Error;
    throw new Error(`Validation failed: ${err.message}`);
  }

  const neo4jSession = context.driver.session();

  let externallyStored: DeleteInput[] = [];
  try {
    return await neo4jSession.writeTransaction(async (tx: Transaction) => {
      const createAliasMappingsResult = await createTypeInstancesInDB(
        tx,
        typeInstancesInput
      );

      const storeInput = getExternallyStoredValues(
        createAliasMappingsResult,
        typeInstancesInput
      );
      externallyStored = storeInput;
      const updatedContexts = await context.delegatedStorage.Store(
        ...storeInput
      );

      await updateTypeInstancesContextInDB(tx, updatedContexts);

      await setTypeInstanceRelationsInDB(
        tx,
        createAliasMappingsResult,
        usesRelations
      );

      return Object.entries(createAliasMappingsResult).map((entry) => ({
        alias: entry[0],
        id: entry[1],
      }));
    });
  } catch (e) {
    // Ensure that data is deleted in case of not committed transaction
    await context.delegatedStorage.Delete(...externallyStored);

    const err = e as Error;
    throw new Error(`failed to create the TypeInstances: ${err.message}`);
  } finally {
    await neo4jSession.close();
  }
}

function validate(input: TypeInstanceInput[]) {
  const aliases = input
    .filter((x) => x.alias !== undefined)
    .map((x) => x.alias);
  if (new Set(aliases).size !== aliases.length) {
    throw new Error(
      "Duplicated TypeInstance aliases. Please ensure that each TypeInstance alias is unique."
    );
  }

  if (input.length !== aliases.length) {
    throw new Error(
      "Missing TypeInstance aliases. Please ensure that each TypeInstance has unique alias."
    );
  }

  const notAllowedBackendCtx = input.filter(
    (x) => x.backend.id === BUILTIN_STORAGE_BACKEND_ID && x.backend.context
  );
  if (notAllowedBackendCtx.length) {
    throw new Error("Built-in storage backend does not accept context");
  }
}

async function createTypeInstancesInDB(
  tx: Transaction,
  typeInstancesInput: CreateTypeInstanceInput[]
): Promise<AliasMapping> {
  try {
    logger.debug(
      "Executing query to create TypeInstance in database",
      typeInstancesInput
    );
    const createTypeInstanceResult = await tx.run(
      `
           // Check if a given backend is registered
           CALL {
             UNWIND $typeInstances AS typeInstance
             OPTIONAL MATCH (backendTI:TypeInstance {id: typeInstance.backend.id})
             WITH backendTI, typeInstance
             WHERE backendTI IS NULL
             RETURN collect(typeInstance.backend.id) as notFoundBackendIds
           }
           CALL apoc.util.validate(size(notFoundBackendIds) > 0, apoc.convert.toJson({ids: notFoundBackendIds, code: 404}), null)
           UNWIND $typeInstances AS typeInstance
           CREATE (ti:TypeInstance {id: apoc.create.uuid(), createdAt: datetime()})

           // Backend
           WITH *
           MATCH (backendTI:TypeInstance {id: typeInstance.backend.id})
           CREATE (ti)-[:USES]->(backendTI)
           // TODO(storage): It should be taken from the uses relation but we don't have access to the TypeRef.additionalRefs to check
           // if a given type is a backend or not. Maybe we will introduce a dedicated property to distinguish them from others.
           MERGE (storageRef:TypeInstanceBackendReference {id: typeInstance.backend.id, abstract: typeInstance.backend.abstract })
           CREATE (ti)-[:STORED_IN]->(storageRef)

					 // TypeRef
           MERGE (typeRef:TypeInstanceTypeReference {path: typeInstance.typeRef.path, revision: typeInstance.typeRef.revision})
           CREATE (ti)-[:OF_TYPE]->(typeRef)

           // Revision
           CREATE (tir: TypeInstanceResourceVersion {resourceVersion: 1, createdBy: typeInstance.createdBy})
           CREATE (ti)-[:CONTAINS]->(tir)

           CREATE (tir)-[:DESCRIBED_BY]->(metadata: TypeInstanceResourceVersionMetadata)
           CREATE (tir)-[:SPECIFIED_BY]->(spec: TypeInstanceResourceVersionSpec {value: apoc.convert.toJson(typeInstance.value)})
           CREATE (specBackend: TypeInstanceResourceVersionSpecBackend {context: apoc.convert.toJson(typeInstance.backend.context)})
           CREATE (spec)-[:WITH_BACKEND]->(specBackend)

           FOREACH (attr in typeInstance.attributes |
             MERGE (attrRef: AttributeReference {path: attr.path, revision: attr.revision})
             CREATE (metadata)-[:CHARACTERIZED_BY]->(attrRef)
           )

           RETURN ti.id as uuid, typeInstance.alias as alias
           `,
      { typeInstances: typeInstancesInput }
    );

    // HINT: returned records may be sometimes duplicated, so we need to reduce them and create an expected mapping upfront.
    const aliasMappings = mapAliasesToIDs(createTypeInstanceResult);
    if (Object.keys(aliasMappings).length !== typeInstancesInput.length) {
      throw new Error(
        "Failed to create some TypeInstances. Please verify, if you provided all the required fields for TypeInstances."
      );
    }

    return aliasMappings;
  } catch (e) {
    let err = e as Error;
    const customErr = tryToExtractCustomCypherError(err);
    if (customErr) {
      switch (customErr.code) {
        case CustomCypherErrorCode.NotFound:
          err = generateNotFoundBackendError(customErr);
          break;
        default:
          err = Error(`Unexpected error code ${customErr.code}`);
          break;
      }
    }
    throw err;
  }
}

function generateNotFoundBackendError(customErr: CustomCypherErrorOutput) {
  if (!Object.prototype.hasOwnProperty.call(customErr, "ids")) {
    // it shouldn't happen
    return Error(`Detected unregistered storage backends`);
  }
  return Error(
    `TypeInstances for storage backends with IDs "${customErr.ids}" were not found`
  );
}

async function updateTypeInstancesContextInDB(
  tx: Transaction,
  updatedContexts: UpdatedContexts
) {
  if (Object.keys(updatedContexts).length) {
    logger.debug("Executing query to update backend contexts");
  }

  await tx.run(
    `
      UNWIND keys($updatedContexts) AS id
      MATCH (ti:TypeInstance {id: id})

      WITH *
      // Get Latest Revision
      CALL {
          WITH ti
          WITH ti
          MATCH (ti)-[:CONTAINS]->(tir:TypeInstanceResourceVersion)
          RETURN tir ORDER BY tir.resourceVersion DESC LIMIT 1
      }

      MATCH (tir)-[:SPECIFIED_BY]->(spec:TypeInstanceResourceVersionSpec)-[:WITH_BACKEND]->(specBackend: TypeInstanceResourceVersionSpecBackend)
      SET specBackend.context = apoc.convert.toJson($updatedContexts[id])

      RETURN ti`,
    { updatedContexts: updatedContexts }
  );
}

function setDefaults(ti: CreateTypeInstanceInput): TypeInstanceInput {
  ti.backend = ti.backend
    ? enrichWithAbstract(ti.backend)
    : builtinStorageBackendDetails();

  // ensure that alias is set, so we can correlate returned ID from Cypher with an input TypeInstance
  // and map it to proper storage backend.
  ti.alias = ti.alias ?? uniqueNamesGenerator(genAdjsColorsAndAnimals);

  return ti as TypeInstanceInput;
}

function enrichWithAbstract(
  backend: TypeInstanceBackendInput
): TypeInstanceBackendDetails {
  return {
    ...backend,
    abstract: backend.id === BUILTIN_STORAGE_BACKEND_ID,
  };
}

function mapAliasesToIDs(createResult: QueryResult): AliasMapping {
  return createResult.records.reduce((acc: { [key: string]: string }, cur) => {
    const uuid = cur.get("uuid");
    const alias = cur.get("alias");

    return {
      ...acc,
      [alias]: uuid,
    };
  }, {});
}

function getExternallyStoredValues(
  aliasMappings: AliasMapping,
  tis: TypeInstanceInput[]
) {
  return tis
    .filter((ti) => ti.backend && ti.alias && !ti.backend.abstract)
    .map((ti) => {
      // We filter out wrong TypeInstances
      // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
      const tiID = aliasMappings[ti.alias!];
      return {
        backend: ti.backend,
        typeInstance: {
          id: tiID,
          value: ti.value,
        },
      } as StoreInput;
    });
}

async function setTypeInstanceRelationsInDB(
  tx: Transaction,
  aliasMappings: AliasMapping,
  usesRelations: TypeInstanceUsesRelationInput[]
) {
  const usesRelationsParams = usesRelations.map(({ from, to }) => ({
    from: aliasMappings[from] || from,
    to: aliasMappings[to] || to,
  }));

  const createRelationsResult = await tx.run(
    `UNWIND $usesRelations as usesRelation
           MATCH (fromTi:TypeInstance {id: usesRelation.from})
           MATCH (toTi:TypeInstance {id: usesRelation.to})
           CREATE (fromTi)-[:USES]->(toTi)
           RETURN fromTi AS from, toTi AS to
           `,
    {
      usesRelations: usesRelationsParams,
    }
  );

  if (createRelationsResult.records.length !== usesRelationsParams.length) {
    throw new Error(
      "Failed to create some relations. Please verify, if you use proper aliases or IDs in relations definition."
    );
  }
}
