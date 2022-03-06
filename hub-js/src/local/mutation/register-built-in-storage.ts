import { Transaction } from "neo4j-driver";
import { BUILTIN_STORAGE_BACKEND_ID } from "../../config";
import { ContextWithDriver } from "./context";
import { TypeInstanceBackendDetails } from "../types/type-instance";

export async function ensureCoreStorageTypeInstance(
  context: ContextWithDriver
) {
  const neo4jSession = context.driver.session();
  const value = {
    acceptValue: false,
    contextSchema: null,
  };
  try {
    await neo4jSession.writeTransaction(async (tx: Transaction) => {
      await tx.run(
        `
            MERGE (ti:TypeInstance {id: $id})
            MERGE (typeRef:TypeInstanceTypeReference {path: "cap.core.type.hub.storage.neo4j", revision: "0.1.0"})
            MERGE (backend:TypeInstanceBackendReference {abstract: true, id: ti.id, description: "Built-in Hub storage"})
            MERGE (tir: TypeInstanceResourceVersion {resourceVersion: 1, createdBy: "core"})
            MERGE (spec: TypeInstanceResourceVersionSpec {value: apoc.convert.toJson($value)})
            MERGE (specBackend: TypeInstanceResourceVersionSpecBackend {context: apoc.convert.toJson(null)})

            MERGE (ti)-[:OF_TYPE]->(typeRef)
            MERGE (ti)-[:STORED_IN]->(backend)
            MERGE (ti)-[:CONTAINS]->(tir)
            MERGE (tir)-[:DESCRIBED_BY]->(metadata:TypeInstanceResourceVersionMetadata)
            MERGE (tir)-[:SPECIFIED_BY]->(spec)
            MERGE (spec)-[:WITH_BACKEND]->(specBackend)

            SET ti.createdAt = CASE WHEN ti.createdAt IS NOT NULL THEN ti.createdAt ELSE datetime() END

            RETURN ti
          `,
        { value, id: BUILTIN_STORAGE_BACKEND_ID }
      );
    });
  } catch (e) {
    const err = e as Error;
    throw new Error(
      `while ensuring TypeInstance for core backend storage: ${err.message}`
    );
  } finally {
    await neo4jSession.close();
  }
}

export function builtinStorageBackendDetails(): TypeInstanceBackendDetails {
  return {
    id: BUILTIN_STORAGE_BACKEND_ID,
    abstract: true,
  };
}
