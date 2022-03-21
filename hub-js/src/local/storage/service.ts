import {
  GetValueRequest,
  OnCreateRequest,
  OnDeleteRequest,
  OnLockRequest,
  OnUnlockRequest,
  OnUpdateRequest,
  StorageBackendDefinition,
} from "../../generated/grpc/storage_backend";
import { Client, createChannel, createClient } from "nice-grpc";
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
import { TextEncoder } from "util";

type StorageClient = Client<typeof StorageBackendDefinition>;

interface BackendContainer {
  client: StorageClient;
  validateSpec: ValidateBackendSpec;
}

interface ValidateBackendSpec {
  backendId: string;
  contextSchema: JSONSchemaType<unknown> | undefined;
  acceptValue: boolean;
}

type ValidateInput = GetInput | UpdateInput | DeleteInput | StoreInput;

export class ValidationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "ValidationError";
  }
}

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
  private registeredClients: Map<string, BackendContainer>;
  private readonly dbDriver: Driver;
  private readonly ajv: Ajv;

  constructor(dbDriver: Driver) {
    this.registeredClients = new Map<string, BackendContainer>();
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
      const backend = await this.getBackendContainer(input.backend.id);

      const validateErr = this.validateInput(input, backend.validateSpec);
      if (validateErr) {
        throw Error(
          `External backend "${input.backend.id}": ${validateErr.message}`
        );
      }

      const req: OnCreateRequest = {
        typeInstanceId: input.typeInstance.id,
        value: DelegatedStorageService.encode(input.typeInstance.value),
        context: DelegatedStorageService.encode(input.backend.context),
      };

      const res = await backend.client.onCreate(req);

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
      const backend = await this.getBackendContainer(input.backend.id);
      const validateErr = this.validateInput(input, backend.validateSpec);
      if (validateErr) {
        throw Error(
          `External backend "${input.backend.id}": ${validateErr.message}`
        );
      }

      const req: OnUpdateRequest = {
        typeInstanceId: input.typeInstance.id,
        newResourceVersion: input.typeInstance.newResourceVersion,
        newValue: DelegatedStorageService.encode(input.typeInstance.newValue),
        context: DelegatedStorageService.encode(input.backend.context),
        ownerId: input.typeInstance.ownerID,
      };

      await backend.client.onUpdate(req);
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
      const backend = await this.getBackendContainer(input.backend.id);

      const validateErr = this.validateInput(input, backend.validateSpec);
      if (validateErr) {
        throw Error(
          `External backend "${input.backend.id}": ${validateErr.message}`
        );
      }

      const req: GetValueRequest = {
        typeInstanceId: input.typeInstance.id,
        resourceVersion: input.typeInstance.resourceVersion,
        context: DelegatedStorageService.encode(input.backend.context),
      };
      const res = await backend.client.getValue(req);

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
      const backend = await this.getBackendContainer(input.backend.id);

      const validateErr = this.validateInput(input, backend.validateSpec);
      if (validateErr) {
        throw Error(
          `External backend "${input.backend.id}": ${validateErr.message}`
        );
      }
      const req: OnDeleteRequest = {
        typeInstanceId: input.typeInstance.id,
        context: DelegatedStorageService.encode(input.backend.context),
        ownerId: input.typeInstance.ownerID,
      };
      await backend.client.onDelete(req);
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
      const backend = await this.getBackendContainer(input.backend.id);

      const validateErr = this.validateInput(input, backend.validateSpec);
      if (validateErr) {
        throw Error(
          `External backend "${input.backend.id}": ${validateErr.message}`
        );
      }

      const req: OnLockRequest = {
        typeInstanceId: input.typeInstance.id,
        lockedBy: input.typeInstance.lockedBy,
        context: DelegatedStorageService.encode(input.backend.context),
      };
      await backend.client.onLock(req);
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
      const backend = await this.getBackendContainer(input.backend.id);

      const validateErr = this.validateInput(input, backend.validateSpec);
      if (validateErr) {
        throw Error(
          `External backend "${input.backend.id}": ${validateErr.message}`
        );
      }

      const req: OnUnlockRequest = {
        typeInstanceId: input.typeInstance.id,
        context: DelegatedStorageService.encode(input.backend.context),
      };
      await backend.client.onUnlock(req);
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

  private async getBackendContainer(id: string): Promise<BackendContainer> {
    if (!this.registeredClients.has(id)) {
      const spec = await this.storageInstanceDetailsFetcher(id);
      logger.debug("Initialize gRPC BackendContainer", {
        backend: id,
        url: spec.url,
      });

      let contextSchema;
      if (spec.contextSchema) {
        const out = DelegatedStorageService.parseToObject(spec.contextSchema);
        if (out.error) {
          throw Error(
            `failed to process the TypeInstance's backend "${id}": invalid spec.context: ${out.error.message}`
          );
        }
        contextSchema = out.parsed as JSONSchemaType<unknown>;
      }
      const channel = createChannel(spec.url);
      const client: StorageClient = createClient(
        StorageBackendDefinition,
        channel
      );

      const storageSpec = {
        backendId: id,
        contextSchema,
        acceptValue: spec.acceptValue,
      };

      this.registeredClients.set(id, { client, validateSpec: storageSpec });
    }

    return this.registeredClients.get(id) as BackendContainer;
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

  private static encode(val: unknown) {
    return new TextEncoder().encode(
      DelegatedStorageService.convertToJSONIfObject(val)
    );
  }

  private static normalizeInput(
    input: GetInput | UpdateInput | DeleteInput | StoreInput
  ) {
    const out: { context?: unknown; value?: unknown } = {
      context: input.backend.context,
      value: undefined,
    };

    if ("value" in input.typeInstance) {
      out.value = input.typeInstance.value;
    }
    if ("newValue" in input.typeInstance) {
      out.value = input.typeInstance.newValue;
    }
    return out;
  }

  private validateInput(
    input: ValidateInput,
    storageSpec: ValidateBackendSpec
  ): ValidationError | undefined {
    const { value, context } = DelegatedStorageService.normalizeInput(input);

    if (!storageSpec.acceptValue && value) {
      return new ValidationError("input value not allowed");
    }

    if (context) {
      if (storageSpec.contextSchema === undefined) {
        return new ValidationError("input context not allowed");
      }

      const validate = this.ajv.compile(storageSpec.contextSchema);
      if (!validate(context)) {
        const msg = this.ajv.errorsText(validate.errors, {
          dataVar: "context",
        });
        return new ValidationError(`invalid input: ${msg}`);
      }
    }

    return undefined;
  }

  private static parseToObject(input: string): {
    error?: Error;
    parsed: unknown;
  } {
    try {
      return {
        parsed: JSON.parse(input),
      };
    } catch (e) {
      const err = e as Error;
      return {
        parsed: {},
        error: err,
      };
    }
  }
}
