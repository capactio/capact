import {
  GetValueRequest,
  OnCreateRequest,
  OnDeleteRequest,
  OnLockRequest,
  OnUnlockRequest,
  OnUpdateRequest,
  StorageBackendDefinition,
} from "../../generated/grpc/storage_backend";
import {
  Client,
  ClientError,
  ClientMiddleware,
  createChannel,
  createClientFactory,
  Status,
} from "nice-grpc";
import { Driver } from "neo4j-driver";
import { TypeInstanceBackendInput } from "../types/type-instance";
import { logger } from "../../logger";
import Ajv from "ajv";
import addFormats from "ajv-formats";
import {
  StorageTypeInstanceSpec,
  StorageTypeInstanceSpecSchema,
} from "./backend-schema";
import { JSONSchemaType } from "ajv/lib/types/json-schema";
import { TextDecoder, TextEncoder } from "util";

// TODO(https://github.com/capactio/capact/issues/634):
// Represents the fake storage backend URL that should be ignored
// as the backend server is not deployed.
// It should be removed after a real backend is used in `test/e2e/action_test.go` scenarios.
export const FAKE_TEST_URL = "e2e-test-backend-mock-url:50051";

type StorageClient = Client<typeof StorageBackendDefinition, DenyOptions>;

export interface StoreInput {
  backend: TypeInstanceBackendInput;
  typeInstance: {
    id: string;
    value: unknown;
  };
}

export interface UpdateInput {
  backend: TypeInstanceBackendInput;
  typeInstance: {
    id: string;
    newResourceVersion: number;
    newValue: unknown;
    ownerID?: string;
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
    ownerID?: string;
  };
}

export interface LockInput {
  backend: TypeInstanceBackendInput;
  typeInstance: {
    id: string;
    lockedBy: string;
  };
}

export interface UnlockInput {
  backend: TypeInstanceBackendInput;
  typeInstance: {
    id: string;
  };
}

export interface UpdatedContexts {
  [key: string]: unknown;
}

export default class DelegatedStorageService {
  private registeredClients: Map<string, StorageClient>;
  private readonly dbDriver: Driver;
  private readonly ajv: Ajv;

  constructor(dbDriver: Driver) {
    this.registeredClients = new Map<string, StorageClient>();
    this.dbDriver = dbDriver;

    this.ajv = new Ajv({ allErrors: true });
    addFormats(this.ajv);
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
      logger.debug("Storing TypeInstance in external backend", {
        typeInstanceId: input.typeInstance.id,
        backendId: input.backend.id,
      });
      const cli = await this.getClient(input.backend.id);
      if (!cli) {
        // TODO: remove after using a real backend in e2e tests.
        continue;
      }

      const req: OnCreateRequest = {
        typeInstanceId: input.typeInstance.id,
        value: DelegatedStorageService.encode(input.typeInstance.value),
        context: DelegatedStorageService.encode(input.backend.context),
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
   * Updates the TypeInstance's value in a given backend.
   *
   *
   * @param inputs - Describes what should be updated.
   *
   */
  async Update(...inputs: UpdateInput[]) {
    for (const input of inputs) {
      logger.debug("Updating TypeInstance in external backend", {
        typeInstanceId: input.typeInstance.id,
        backendId: input.backend.id,
      });
      const cli = await this.getClient(input.backend.id);
      if (!cli) {
        // TODO(https://github.com/capactio/capact/issues/634): remove after using a real backend in e2e tests.
        continue;
      }

      const req: OnUpdateRequest = {
        typeInstanceId: input.typeInstance.id,
        newResourceVersion: input.typeInstance.newResourceVersion,
        newValue: DelegatedStorageService.encode(input.typeInstance.newValue),
        context: DelegatedStorageService.encode(input.backend.context),
        ownerId: input.typeInstance.ownerID,
      };

      await cli.onUpdate(req);
    }
  }

  /**
   * Gets the TypeInstance's value from a given backend.
   *
   *
   * @param inputs - Describes what should be stored.
   * @returns The update backend's context. If there was no update, it's undefined.
   *
   */
  async Get(...inputs: GetInput[]): Promise<UpdatedContexts> {
    let result: UpdatedContexts = {};

    for (const input of inputs) {
      logger.debug("Fetching TypeInstance from external backend", {
        typeInstanceId: input.typeInstance.id,
        backendId: input.backend.id,
      });
      const cli = await this.getClient(input.backend.id);
      if (!cli) {
        // TODO(https://github.com/capactio/capact/issues/634): remove after using a real backend in e2e tests.
        result = {
          ...result,
          [input.typeInstance.id]: {
            key: input.backend.id,
          },
        };
        continue;
      }

      const req: GetValueRequest = {
        typeInstanceId: input.typeInstance.id,
        resourceVersion: input.typeInstance.resourceVersion,
        context: DelegatedStorageService.encode(input.backend.context),
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
   * Deletes a given TypeInstance
   *
   * @param inputs - Describes what should be deleted.
   *
   */
  async Delete(...inputs: DeleteInput[]) {
    for (const input of inputs) {
      logger.debug("Deleting TypeInstance from external backend", {
        typeInstanceId: input.typeInstance.id,
        backendId: input.backend.id,
      });
      const cli = await this.getClient(input.backend.id);
      if (!cli) {
        // TODO(https://github.com/capactio/capact/issues/634): remove after using a real backend in e2e tests.
        continue;
      }

      const req: OnDeleteRequest = {
        typeInstanceId: input.typeInstance.id,
        context: DelegatedStorageService.encode(input.backend.context),
        ownerId: input.typeInstance.ownerID,
      };
      await cli.onDelete(req);
    }
  }

  /**
   * Locks a given TypeInstance
   *
   * @param inputs - Describes what should be locked. Owner ID is needed.
   *
   */
  async Lock(...inputs: LockInput[]) {
    for (const input of inputs) {
      logger.debug("Locking TypeInstance in external backend", {
        typeInstanceId: input.typeInstance.id,
        backendId: input.backend.id,
      });
      const cli = await this.getClient(input.backend.id);
      if (!cli) {
        // TODO(https://github.com/capactio/capact/issues/634): remove after using a real backend in e2e tests.
        continue;
      }

      const req: OnLockRequest = {
        typeInstanceId: input.typeInstance.id,
        lockedBy: input.typeInstance.lockedBy,
        context: DelegatedStorageService.encode(input.backend.context),
      };
      await cli.onLock(req);
    }
  }

  /**
   * Unlocks a given TypeInstance
   *
   * @param inputs - Describes what should be unlocked. Owner ID is not needed.
   *
   */
  async Unlock(...inputs: UnlockInput[]) {
    for (const input of inputs) {
      logger.debug(`Unlocking TypeInstance in external backend`, {
        typeInstanceId: input.typeInstance.id,
        backendId: input.backend.id,
      });
      const cli = await this.getClient(input.backend.id);
      if (!cli) {
        // TODO(https://github.com/capactio/capact/issues/634): remove after using a real backend in e2e tests.
        continue;
      }

      const req: OnUnlockRequest = {
        typeInstanceId: input.typeInstance.id,
        context: DelegatedStorageService.encode(input.backend.context),
      };
      await cli.onUnlock(req);
    }
  }

  private async storageInstanceDetailsFetcher(
    id: string
  ): Promise<StorageTypeInstanceSpec> {
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

      const storageSpec: StorageTypeInstanceSpec = record.get("value");

      this.validateStorageSpecValue(storageSpec);

      return storageSpec;
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
      const spec = await this.storageInstanceDetailsFetcher(id);
      if (spec.url === FAKE_TEST_URL) {
        logger.debug(
          "Skipping a real call as backend was classified as a fake one"
        );
        // TODO(https://github.com/capactio/capact/issues/634): remove after using a real backend in e2e tests.
        return undefined;
      }

      logger.debug("Initialize gRPC client", {
        backend: id,
        url: spec.url,
      });

      const clientFactory = createClientFactory().use(
        newValidateMiddleware(this.ajv)
      );

      let contextSchema = null;
      if (spec.contextSchema) {
        contextSchema = JSON.parse(spec.contextSchema) as JSONSchemaType<{
          [key: string]: any;
        }>;
      }
      const channel = createChannel(spec.url);
      const client: StorageClient = clientFactory.create(
        StorageBackendDefinition,
        channel,
        {
          "*": {
            storageSpec: {
              contextSchema,
              acceptValue: spec.acceptValue,
            },
          },
        }
      );
      this.registeredClients.set(id, client);
    }

    return this.registeredClients.get(id);
  }

  private static convertToJSONIfObject(val: unknown): string | undefined {
    if (val instanceof Array || typeof val === "object") {
      return JSON.stringify(val);
    }
    return val as string;
  }

  private encode(val: unknown) {
    return new TextEncoder().encode(
      DelegatedStorageService.convertToJSONIfObject(val)
    );
  }
  private validateStorageSpecValue(storageSpec: StorageTypeInstanceSpec) {
    const validate = this.ajv.compile(StorageTypeInstanceSpecSchema);

    if (validate(storageSpec)) {
      return;
    }

    throw new Error(
      this.ajv.errorsText(validate.errors, { dataVar: "spec.value" })
    );
  }
}

interface ContextJSONSchema {
  [key: string]: any;
}

interface DenyOptions {
  storageSpec?: {
    contextSchema: JSONSchemaType<ContextJSONSchema> | null;
    acceptValue: boolean;
  };
}

function newValidateMiddleware(ajv: Ajv): ClientMiddleware<DenyOptions> {
  return async function* denyMiddleware(call, options) {
    if (!options.storageSpec) {
      return yield* call.next(call.request, options);
    }

    const { storageSpec, ...restOptions } = options;
    const hasValue = Object.prototype.hasOwnProperty.call(
      call.request,
      "value"
    );
    const hasContext = Object.prototype.hasOwnProperty.call(
      call.request,
      "context"
    );

    if (!options.storageSpec?.acceptValue && hasValue) {
      throw new ClientError(
        call.method.path,
        Status.INVALID_ARGUMENT,
        "Delegated storage doesn't accept value"
      );
    }

    if (hasContext) {
      if (options.storageSpec.contextSchema === null) {
        throw new ClientError(
          call.method.path,
          Status.INVALID_ARGUMENT,
          "Delegated storage doesn't accept context"
        );
      }

      console.log(options.storageSpec.contextSchema);
      const validate = ajv.compile(options.storageSpec.contextSchema);

      // @ts-ignore
      const ctx = new TextDecoder().decode(call.request.context);
      const ctxObj: ContextJSONSchema = JSON.parse(ctx);
      console.log(ctxObj);
      console.log(typeof ctxObj);
      console.log(typeof options.storageSpec.contextSchema);
      if (validate(ctxObj)) {
      } else {
        throw new ClientError(
          call.method.path,
          Status.INVALID_ARGUMENT,
          ajv.errorsText(validate.errors, { dataVar: "context" })
        );
      }
    }
    return yield* call.next(call.request, {
      ...restOptions,
    });
  };
}
