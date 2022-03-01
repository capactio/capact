import { QueryResult, Transaction } from "neo4j-driver";
import { BUILTIN_STORAGE_BACKEND_ID } from "../../config";
import { Context } from "./context";
import DelegatedStorageService from "../storage/service";
import {
  CreateTypeInstanceInput,
  CreateTypeInstancesInput,
  TypeInstanceBackendDetails,
  TypeInstanceUsesRelationInput,
} from "../types/type-instance";

interface createTypeInstancesArgs {
  in: CreateTypeInstancesInput;
}

interface aliasMapping {
  [key: string]: string;
}

export async function createTypeInstances(
  _: any,
  args: createTypeInstancesArgs,
  context: Context
) {
  const { typeInstances, usesRelations } = args.in;
  try {
    validate(typeInstances);
  } catch (e) {
    const err = e as Error;
    throw new Error(`Validation failed: ${err.message}`);
  }

  const neo4jSession = context.driver.session();

  try {
    return await neo4jSession.writeTransaction(async (tx: Transaction) => {
      const typeInstancesInput = typeInstances.map(setStorageBackend);
      const createTypeInstanceResult = await createTypeInstancesInDB(
        tx,
        typeInstancesInput
      );

      const aliasMappings = mapAliasesToIDs(createTypeInstanceResult);

      console.log("here2");
      await delegatedDataStorageIfNeeded(
        context.delegatedStorage,
        aliasMappings,
        typeInstancesInput
      );
      console.log("here3");
      await setTypeInstanceRelationsInDB(tx, aliasMappings, usesRelations);

      // TODO: update context if needed.

      return Object.entries(aliasMappings).map((entry) => ({
        alias: entry[0],
        id: entry[1],
      }));
    });
  } catch (e) {
    // for _ := ti { ensure data is deleted in delegated storage }
    const err = e as Error;
    throw new Error(`failed to create the TypeInstances: ${err.message}`);
  } finally {
    await neo4jSession.close();
  }
}

function validate(typeInstances: CreateTypeInstanceInput[]) {
  const aliases = typeInstances
    .filter((x) => x.alias !== undefined)
    .map((x) => x.alias);
  if (new Set(aliases).size !== aliases.length) {
    throw new Error(
      "Duplicated TypeInstance aliases. Please ensure that each TypeInstance alias is unique."
    );
  }
  if (typeInstances.length !== aliases.length) {
    throw new Error(
      "Missing TypeInstance aliases. Please ensure that each TypeInstance has unique alias."
    );
  }
}

async function createTypeInstancesInDB(
  tx: Transaction,
  typeInstancesInput: CreateTypeInstanceInput[]
) {
  const createTypeInstanceResult = await tx.run(
    `UNWIND $typeInstances AS typeInstance
           CREATE (ti:TypeInstance {id: apoc.create.uuid(), createdAt: datetime()})

           // Backend
           WITH *
           MATCH (backendTI:TypeInstance {id: typeInstance.backend.id})
           CREATE (ti)-[:USES]->(backendTI)
           // TODO(storage): It should be taken from the uses relation but we don't have access to the TypeRef.additionalRefs to check
           // if a given type is a backend or not. Maybe we will introduce a dedicated property to distinguish them from others.
           MERGE (storageRef:TypeInstanceBackendReference)
           SET storageRef = typeInstance.backend
           CREATE (ti)-[:STORED_IN]->(storageRef)

					 // TypeRef
           MERGE (typeRef:TypeInstanceTypeReference {path: typeInstance.typeRef.path, revision: typeInstance.typeRef.revision})
           CREATE (ti)-[:OF_TYPE]->(typeRef)

           // Revision
           CREATE (tir: TypeInstanceResourceVersion {resourceVersion: 1, createdBy: typeInstance.createdBy})
           CREATE (ti)-[:CONTAINS]->(tir)

           CREATE (tir)-[:DESCRIBED_BY]->(metadata: TypeInstanceResourceVersionMetadata)
           CREATE (tir)-[:SPECIFIED_BY]->(spec: TypeInstanceResourceVersionSpec {value: apoc.convert.toJson(typeInstance.value)})

           FOREACH (attr in typeInstance.attributes |
             MERGE (attrRef: AttributeReference {path: attr.path, revision: attr.revision})
             CREATE (metadata)-[:CHARACTERIZED_BY]->(attrRef)
           )

           RETURN ti.id as uuid, typeInstance.alias as alias
           `,
    { typeInstances: typeInstancesInput }
  );

  if (createTypeInstanceResult.records.length !== typeInstancesInput.length) {
    throw new Error(
      "Failed to create some TypeInstances. Please verify, if you provided all the required fields for TypeInstances."
    );
  }

  return createTypeInstanceResult;
}

export type TypeInstanceInput = Omit<CreateTypeInstanceInput, "backend"> & {
  backend: TypeInstanceBackendDetails;
};

function setStorageBackend(ti: CreateTypeInstanceInput): TypeInstanceInput {
  ti.backend = {
    id: ti.backend?.id || BUILTIN_STORAGE_BACKEND_ID, // if not provided, store in built-in
    abstract: !ti.backend,
  } as TypeInstanceBackendDetails;

  return ti as TypeInstanceInput;
}

function mapAliasesToIDs(createResult: QueryResult): aliasMapping {
  return createResult.records.reduce((acc: { [key: string]: string }, cur) => {
    const uuid = cur.get("uuid");
    const alias = cur.get("alias");

    return {
      ...acc,
      [alias]: uuid,
    };
  }, {});
}

async function delegatedDataStorageIfNeeded(
  delegatedStorage: DelegatedStorageService,
  aliasMappings: aliasMapping,
  tis: TypeInstanceInput[]
) {
  for (const ti of tis) {
    if (!ti.backend || !ti.alias) {
      continue;
    }

    if (ti.backend.abstract) {
      // TODO: can be helper method in backend
      continue;
    }
    console.info(ti);
    console.info(ti.value);

    await delegatedStorage.Store({
      backend: ti.backend,
      typeInstance: {
        id: aliasMappings[ti.alias],
        value: ti.value,
      },
    });
  }
}

async function setTypeInstanceRelationsInDB(
  tx: Transaction,
  aliasMappings: aliasMapping,
  usesRelations: TypeInstanceUsesRelationInput[]
) {
  const usesRelationsParams = usesRelations.map(
    ({ from, to }: { from: string; to: string }) => ({
      from: aliasMappings[from] || from,
      to: aliasMappings[to] || to,
    })
  );

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
