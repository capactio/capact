import {
  OnCreateRequest,
  StorageBackendDefinition
} from "../../../grpc/storage_backend";
import { createChannel, createClient, Client } from "nice-grpc";
import { Driver } from "neo4j-driver";
import { TypeInstanceBackendInput } from "../types/type-instance";

export const TARGET = "0.0.0.0:50051";

type StorageClient = Client<typeof StorageBackendDefinition>

export interface StorageInstanceDetails {
  url: string;
}

export interface StoreInput {
  backend: TypeInstanceBackendInput,
  typeInstance: {
    id: string
    value: any;
  };
}

export default class DelegatedStorageService {
  private registeredClients: Map<string, StorageClient>;
  private dbDriver: Driver;

  constructor(dbDriver: Driver) {
    this.registeredClients = new Map<string, StorageClient>();
    this.dbDriver = dbDriver;
  }

  /**
   * Stores the TypeInstance's value in a given backend.
   *
   *
   * @param input - Describes what should be stored.
   * @returns The update backend's context. If there was no update, it's undefined.
   *
   */
  async Store(input: StoreInput): Promise<Uint8Array | undefined> {
    const req: OnCreateRequest = {
      typeInstanceId: input.typeInstance.id,
      value: new TextEncoder().encode(JSON.stringify(input.typeInstance.value))
      // TODO: will be done in follow-up pull-request for #645
      // context: input.backend.context,
    };
    const cli = await this.getClient(input.backend.id);
    const res = await cli.onCreate(req);

    return res.context;
  }

  private async storageInstanceDetailsFetcher(id: string): Promise<StorageInstanceDetails> {
    const sess = this.dbDriver.session();
    try {
      const fetchRevisionResult = await sess.run(
        `
           MATCH (ti:TypeInstance {id: $id})
           WITH *
           CALL {
             WITH ti
             MATCH (ti)-[:CONTAINS]->(tir:TypeInstanceResourceVersion)
             RETURN tir ORDER BY tir.resourceVersion DESC LIMIT 1
           }
           MATCH (tir)-[:SPECIFIED_BY]->(spec:TypeInstanceResourceVersionSpec)
           RETURN apoc.convert.fromJsonMap(spec.value) as value
          `,
        { id: id }
      );
      if (fetchRevisionResult.records.length !== 1) {
        throw new Error(`Internal Server Error, unexpected response row length, want 1, got ${fetchRevisionResult.records.length}`);
      }

      const record = fetchRevisionResult.records[0];
      return record.get("value"); // TODO: validate against Storage JSON Schema.

    } catch (e) {
      const err = e as Error;
      throw new Error(`failed to fetch the TypeInstance "${id}": ${err.message}`);
    } finally {
      await sess.close();
    }
  }

  private async getClient(id: string): Promise<StorageClient> {
    if (!this.registeredClients.has(id)) {
      const { url } = await this.storageInstanceDetailsFetcher(id);
      const channel = createChannel(url);
      const client: StorageClient = createClient(
        StorageBackendDefinition,
        channel
      );
      this.registeredClients.set(id, client);
    }
    return this.registeredClients.get(id)!;
  }
}
