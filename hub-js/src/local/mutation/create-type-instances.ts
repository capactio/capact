import { QueryResult, Transaction } from "neo4j-driver";
import { BUILTIN_STORAGE_BACKEND_ID } from "../../config";
import { Context } from "./context";
import {
  uniqueNamesGenerator,
  Config,
  adjectives,
  colors,
  animals
} from "unique-names-generator";
import {
  DeleteInput,
  StoreInput,
  UpdatedContexts
} from "../storage/service";
import {
  CreateTypeInstanceInput,
  CreateTypeInstancesInput,
  TypeInstanceBackendDetails,
  TypeInstanceUsesRelationInput
} from "../types/type-instance";

const genAdjsColorsAndAnimals: Config = {
  dictionaries: [adjectives, colors, animals],
  separator: "_",
  length: 3
};

interface createTypeInstancesArgs {
  in: CreateTypeInstancesInput;
}

interface aliasMapping {
  [key: string]: string;
}

export type TypeInstanceInput = Omit<CreateTypeInstanceInput, "backend"> & {
  backend: TypeInstanceBackendDetails;
};

export async function createTypeInstances(
  _: any,
  args: createTypeInstancesArgs,
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

      const storeInput = getExternallyStoredValues(createAliasMappingsResult, typeInstancesInput);
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
        id: entry[1]
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
}

async function createTypeInstancesInDB(
  tx: Transaction,
  typeInstancesInput: CreateTypeInstanceInput[]
): Promise<aliasMapping> {
  const createTypeInstanceResult = await tx.run(
    `UNWIND $typeInstances AS typeInstance
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
}

async function updateTypeInstancesContextInDB(
  tx: Transaction,
  updatedContexts: UpdatedContexts
) {
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
  ti.backend = {
    id: ti.backend?.id || BUILTIN_STORAGE_BACKEND_ID, // if not provided, store in built-in
    abstract: !ti.backend,
    context: ti.backend?.context
  } as TypeInstanceBackendDetails;

  // ensure that alias is set, so we can correlate returned ID from Cypher with an input TypeInstance
  // and map it to proper storage backend.
  ti.alias = ti.alias ?? uniqueNamesGenerator(genAdjsColorsAndAnimals);

  return ti as TypeInstanceInput;
}

function mapAliasesToIDs(createResult: QueryResult): aliasMapping {
  return createResult.records.reduce((acc: { [key: string]: string }, cur) => {
    const uuid = cur.get("uuid");
    const alias = cur.get("alias");

    return {
      ...acc,
      [alias]: uuid
    };
  }, {});
}

function getExternallyStoredValues(aliasMappings: aliasMapping, tis: TypeInstanceInput[]) {
  return tis
    .filter((ti) => ti.backend && ti.alias && !ti.backend.abstract)
    .map((ti) => {
      const tiID = aliasMappings[ti.alias!];
      return {
        backend: ti.backend,
        typeInstance: {
          id: tiID,
          value: ti.value
        }
      } as StoreInput;
    });
}

async function setTypeInstanceRelationsInDB(
  tx: Transaction,
  aliasMappings: aliasMapping,
  usesRelations: TypeInstanceUsesRelationInput[]
) {
  const usesRelationsParams = usesRelations.map(({ from, to }) => ({
    from: aliasMappings[from] || from,
    to: aliasMappings[to] || to
  }));

  const createRelationsResult = await tx.run(
    `UNWIND $usesRelations as usesRelation
           MATCH (fromTi:TypeInstance {id: usesRelation.from})
           MATCH (toTi:TypeInstance {id: usesRelation.to})
           CREATE (fromTi)-[:USES]->(toTi)
           RETURN fromTi AS from, toTi AS to
           `,
    {
      usesRelations: usesRelationsParams
    }
  );

  if (createRelationsResult.records.length !== usesRelationsParams.length) {
    throw new Error(
      "Failed to create some relations. Please verify, if you use proper aliases or IDs in relations definition."
    );
  }
}
