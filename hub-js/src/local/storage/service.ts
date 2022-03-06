import {
  GetValueRequest,
  OnCreateRequest,
  OnDeleteRequest,
  OnUpdateRequest,
  StorageBackendDefinition,
} from "../../generated/grpc/storage_backend";
import { createChannel, createClient, Client } from "nice-grpc";
import { Driver } from "neo4j-driver";
import { TypeInstanceBackendInput } from "../types/type-instance";
import { logger } from "../../logger";

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
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    value: any;
  };
}

export interface UpdateInput {
  backend: TypeInstanceBackendInput;
  typeInstance: {
    id: string;
    newResourceVersion: number;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    newValue: any;
  };
}

export interface GetInput {
  backend: TypeInstanceBackendInput;
  typeInstance: {
    id: string;
    resourceVersion: number;
  };
}

export interface DeleteInput {
  backend: TypeInstanceBackendInput;
  typeInstance: {
    id: string;
  };
}

export interface UpdatedContexts {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
   *  TODO: validate if `input.value` is allowed by backend (`backend.acceptValue`)
   *  TODO: validate `input.backend.context` against `backend.contextSchema`.
   */
  async Store(...inputs: StoreInput[]): Promise<UpdatedContexts> {
    let mapping: UpdatedContexts = {};

    for (const input of inputs) {
      logger.debug(
        `Storing TypeInstance ${input.typeInstance.id} in external backend ${input.backend.id}`
      );
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
        context: this.encode(input.backend.context),
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
   * Update the TypeInstance's value in a given backend.
   *
   *
   * @param inputs - Describes what should be updated.
   *
   *  TODO: validate if `input.value` is allowed by backend (`backend.acceptValue`)
   *  TODO: validate `input.backend.context` against `backend.contextSchema`.
   */
  async Update(...inputs: UpdateInput[]) {
    for (const input of inputs) {
      logger.debug(
        `Updating TypeInstance ${input.typeInstance.id} in external backend ${input.backend.id}`
      );
      const cli = await this.getClient(input.backend.id);
      if (!cli) {
        // TODO: remove after using a real backend in e2e tests.
        continue;
      }

      const req: OnUpdateRequest = {
        typeInstanceId: input.typeInstance.id,
        newResourceVersion: input.typeInstance.newResourceVersion,
        newValue: new TextEncoder().encode(
          JSON.stringify(input.typeInstance.newValue)
        ),
        context: this.encode(input.backend.context),
      };

      await cli.onUpdate(req);
    }
  }

  /**
   * Get the TypeInstance's value from a given backend.
   *
   *
   * @param inputs - Describes what should be stored.
   * @returns The update backend's context. If there was no update, it's undefined.
   *
   */
  async Get(...inputs: GetInput[]): Promise<UpdatedContexts> {
    let result: UpdatedContexts = {};

    for (const input of inputs) {
      logger.debug(
        `Fetching TypeInstance ${input.typeInstance.id} from external backend ${input.backend.id}`
      );
      const cli = await this.getClient(input.backend.id);
      if (!cli) {
        // TODO: remove after using a real backend in e2e tests.
        continue;
      }

      const req: GetValueRequest = {
        typeInstanceId: input.typeInstance.id,
        resourceVersion: input.typeInstance.resourceVersion,
        context: this.encode(input.backend.context),
      };
      const res = await cli.getValue(req);

      if (!res.value) {
        throw Error(
          `Got empty response for TypeInstance ${input.typeInstance.id} from external backend ${input.backend.id}`
        );
      }

      const decodeRes = JSON.parse(res.value.toString());
      result = {
        ...result,
        [input.typeInstance.id]: decodeRes,
      };
    }

    return result;
  }

  /**
   * Delete a given TypeInstance
   *
   * @param inputs - Describes what should be deleted.
   *
   */
  async Delete(...inputs: DeleteInput[]) {
    for (const input of inputs) {
      logger.debug(
        `Deleting TypeInstance ${input.typeInstance.id} from external backend ${input.backend.id}`
      );
      const cli = await this.getClient(input.backend.id);
      if (!cli) {
        // TODO: remove after using a real backend in e2e tests.
        continue;
      }

      const req: OnDeleteRequest = {
        typeInstanceId: input.typeInstance.id,
        context: new TextEncoder().encode(
          DelegatedStorageService.convertToJSONIfObject(input.backend.context)
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
      switch (fetchRevisionResult.records.length) {
        case 0:
          throw new Error(`TypeInstance not found`);
        case 1:
          break;
        default:
          throw new Error(
            `Found ${fetchRevisionResult.records.length} TypeInstances with the same id`
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
        logger.debug(
          "Skipping a real call as backend was classified as a fake one"
        );
        // TODO: remove after using a real backend in e2e tests.
        return undefined;
      }

      logger.debug(`Initialize gRPC client for Backend ${id} with URL ${url}`);
      const channel = createChannel(url);
      const client: StorageClient = createClient(
        StorageBackendDefinition,
        channel
      );
      this.registeredClients.set(id, client);
    }

    return this.registeredClients.get(id);
  }

  private static convertToJSONIfObject(val: any) {
    if (val instanceof Array || val instanceof Object) {
      return JSON.stringify(val);
    }
    return val;
  }

  private encode(val: any) {
    return new TextEncoder().encode(
      DelegatedStorageService.convertToJSONIfObject(val)
    );
  }
}