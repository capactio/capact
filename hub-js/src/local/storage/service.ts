import {
  OnCreateRequest,
  OnDeleteRequest,
  StorageBackendDefinition,
} from "../../../grpc/storage_backend";
import { createChannel, createClient, Client } from "nice-grpc";
import { Driver } from "neo4j-driver";
import { TypeInstanceBackendInput } from "../types/type-instance";

// TODO(https://github.com/capactio/capact/issues/604):
// Represents the fake storage backend URL that should be ignored
// as the backend server is not deployed.
// It should be removed after a real backend is used in `test/e2e/action_test.go` scenarios.
export const FAKE_TEST_URL = "e2e-test-backend-mock-url:50051";

type StorageClient = Client<typeof StorageBackendDefinition>;

export interface StorageInstanceDetails {
  url: string;
}

export interface StoreInput {
  backend: TypeInstanceBackendInput;
  typeInstance: {
    id: string;
    value: any;
  };
}

export interface DeleteInput {
  backend: TypeInstanceBackendInput;
  typeInstance: {
    id: string;
  };
}

export interface UpdatedContexts {
  [key: string]: any;
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
   * @param inputs - Describes what should be stored.
   * @returns The update backend's context. If there was no update, it's undefined.
   *
   */
  async Store(...inputs: StoreInput[]): Promise<UpdatedContexts> {
    let mapping: UpdatedContexts = {};

    for (const input of inputs) {
      const cli = await this.getClient(input.backend.id);
      if (!cli) {
        // TODO: remove after using a real backend in e2e tests.
        continue;
      }

      const req: OnCreateRequest = {
        typeInstanceId: input.typeInstance.id,
        value: new TextEncoder().encode(
          JSON.stringify(input.typeInstance.value)
        ),
        context: new TextEncoder().encode(
          JSON.stringify(input.backend.context)
        ),
      };
      const res = await cli.onCreate(req);

      if (!res.context) {
        continue;
      }

      const updateCtx = JSON.parse(res.context.toString());
      mapping = {
        ...mapping,
        [input.typeInstance.id]: updateCtx,
      };
    }

    return mapping;
  }

  /**
   * Delete a given TypeInstance
   *
   * @param inputs - Describes what should be deleted.
   *
   */
  async Delete(...inputs: DeleteInput[]) {
    for (const input of inputs) {
      const cli = await this.getClient(input.backend.id);
      if (!cli) {
        // TODO: remove after using a real backend in e2e tests.
        continue;
      }

      const req: OnDeleteRequest = {
        typeInstanceId: input.typeInstance.id,
        context: new TextEncoder().encode(
          JSON.stringify(input.backend.context)
        ),
      };
      await cli.onDelete(req);
    }
  }

  private async storageInstanceDetailsFetcher(
    id: string
  ): Promise<StorageInstanceDetails> {
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
        throw new Error(
          `Internal Server Error, unexpected response row length, want 1, got ${fetchRevisionResult.records.length}`
        );
      }

      const record = fetchRevisionResult.records[0];
      return record.get("value"); // TODO: validate against Storage JSON Schema.
    } catch (e) {
      const err = e as Error;
      throw new Error(
        `failed to resolve the TypeInstance's backend "${id}": ${err.message}`
      );
    } finally {
      await sess.close();
    }
  }

  private async getClient(id: string): Promise<StorageClient | undefined> {
    if (!this.registeredClients.has(id)) {
      const { url } = await this.storageInstanceDetailsFetcher(id);
      if (url === FAKE_TEST_URL) {
        // TODO: remove after using a real backend in e2e tests.
        return undefined;
      }
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
