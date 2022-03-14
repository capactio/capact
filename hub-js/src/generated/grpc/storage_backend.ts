/* eslint-disable */
import Long from "long";
import _m0 from "protobufjs/minimal";

export const protobufPackage = "storage_backend";

export interface OnCreateRequest {
  typeInstanceId: string;
  value: Uint8Array;
  context?: Uint8Array | undefined;
}

export interface OnCreateResponse {
  context?: Uint8Array | undefined;
}

export interface TypeInstanceResourceVersion {
  resourceVersion: number;
  value: Uint8Array;
}

export interface OnUpdateRequest {
  typeInstanceId: string;
  newResourceVersion: number;
  newValue: Uint8Array;
  context?: Uint8Array | undefined;
  ownerId?: string | undefined;
}

export interface OnUpdateResponse {
  context?: Uint8Array | undefined;
}

export interface OnDeleteRequest {
  typeInstanceId: string;
  context?: Uint8Array | undefined;
  ownerId?: string | undefined;
}

export interface OnDeleteResponse {}

export interface GetValueRequest {
  typeInstanceId: string;
  resourceVersion: number;
  context: Uint8Array;
}

export interface GetValueResponse {
  value?: Uint8Array | undefined;
}

export interface GetLockedByRequest {
  typeInstanceId: string;
  context: Uint8Array;
}

export interface GetLockedByResponse {
  lockedBy?: string | undefined;
}

export interface OnLockRequest {
  typeInstanceId: string;
  context: Uint8Array;
  lockedBy: string;
}

export interface OnLockResponse {}

export interface OnUnlockRequest {
  typeInstanceId: string;
  context: Uint8Array;
}

export interface OnUnlockResponse {}

function createBaseOnCreateRequest(): OnCreateRequest {
  return { typeInstanceId: "", value: new Uint8Array(), context: undefined };
}

export const OnCreateRequest = {
  encode(
    message: OnCreateRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.typeInstanceId !== "") {
      writer.uint32(10).string(message.typeInstanceId);
    }
    if (message.value.length !== 0) {
      writer.uint32(18).bytes(message.value);
    }
    if (message.context !== undefined) {
      writer.uint32(26).bytes(message.context);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OnCreateRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseOnCreateRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.typeInstanceId = reader.string();
          break;
        case 2:
          message.value = reader.bytes();
          break;
        case 3:
          message.context = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): OnCreateRequest {
    return {
      typeInstanceId: isSet(object.typeInstanceId)
        ? String(object.typeInstanceId)
        : "",
      value: isSet(object.value)
        ? bytesFromBase64(object.value)
        : new Uint8Array(),
      context: isSet(object.context)
        ? bytesFromBase64(object.context)
        : undefined,
    };
  },

  toJSON(message: OnCreateRequest): unknown {
    const obj: any = {};
    message.typeInstanceId !== undefined &&
      (obj.typeInstanceId = message.typeInstanceId);
    message.value !== undefined &&
      (obj.value = base64FromBytes(
        message.value !== undefined ? message.value : new Uint8Array()
      ));
    message.context !== undefined &&
      (obj.context =
        message.context !== undefined
          ? base64FromBytes(message.context)
          : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<OnCreateRequest>): OnCreateRequest {
    const message = createBaseOnCreateRequest();
    message.typeInstanceId = object.typeInstanceId ?? "";
    message.value = object.value ?? new Uint8Array();
    message.context = object.context ?? undefined;
    return message;
  },
};

function createBaseOnCreateResponse(): OnCreateResponse {
  return { context: undefined };
}

export const OnCreateResponse = {
  encode(
    message: OnCreateResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.context !== undefined) {
      writer.uint32(10).bytes(message.context);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OnCreateResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseOnCreateResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): OnCreateResponse {
    return {
      context: isSet(object.context)
        ? bytesFromBase64(object.context)
        : undefined,
    };
  },

  toJSON(message: OnCreateResponse): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context =
        message.context !== undefined
          ? base64FromBytes(message.context)
          : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<OnCreateResponse>): OnCreateResponse {
    const message = createBaseOnCreateResponse();
    message.context = object.context ?? undefined;
    return message;
  },
};

function createBaseTypeInstanceResourceVersion(): TypeInstanceResourceVersion {
  return { resourceVersion: 0, value: new Uint8Array() };
}

export const TypeInstanceResourceVersion = {
  encode(
    message: TypeInstanceResourceVersion,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.resourceVersion !== 0) {
      writer.uint32(8).uint32(message.resourceVersion);
    }
    if (message.value.length !== 0) {
      writer.uint32(18).bytes(message.value);
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): TypeInstanceResourceVersion {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseTypeInstanceResourceVersion();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.resourceVersion = reader.uint32();
          break;
        case 2:
          message.value = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): TypeInstanceResourceVersion {
    return {
      resourceVersion: isSet(object.resourceVersion)
        ? Number(object.resourceVersion)
        : 0,
      value: isSet(object.value)
        ? bytesFromBase64(object.value)
        : new Uint8Array(),
    };
  },

  toJSON(message: TypeInstanceResourceVersion): unknown {
    const obj: any = {};
    message.resourceVersion !== undefined &&
      (obj.resourceVersion = Math.round(message.resourceVersion));
    message.value !== undefined &&
      (obj.value = base64FromBytes(
        message.value !== undefined ? message.value : new Uint8Array()
      ));
    return obj;
  },

  fromPartial(
    object: DeepPartial<TypeInstanceResourceVersion>
  ): TypeInstanceResourceVersion {
    const message = createBaseTypeInstanceResourceVersion();
    message.resourceVersion = object.resourceVersion ?? 0;
    message.value = object.value ?? new Uint8Array();
    return message;
  },
};

function createBaseOnUpdateRequest(): OnUpdateRequest {
  return {
    typeInstanceId: "",
    newResourceVersion: 0,
    newValue: new Uint8Array(),
    context: undefined,
    ownerId: undefined,
  };
}

export const OnUpdateRequest = {
  encode(
    message: OnUpdateRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.typeInstanceId !== "") {
      writer.uint32(10).string(message.typeInstanceId);
    }
    if (message.newResourceVersion !== 0) {
      writer.uint32(16).uint32(message.newResourceVersion);
    }
    if (message.newValue.length !== 0) {
      writer.uint32(26).bytes(message.newValue);
    }
    if (message.context !== undefined) {
      writer.uint32(34).bytes(message.context);
    }
    if (message.ownerId !== undefined) {
      writer.uint32(42).string(message.ownerId);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OnUpdateRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseOnUpdateRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.typeInstanceId = reader.string();
          break;
        case 2:
          message.newResourceVersion = reader.uint32();
          break;
        case 3:
          message.newValue = reader.bytes();
          break;
        case 4:
          message.context = reader.bytes();
          break;
        case 5:
          message.ownerId = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): OnUpdateRequest {
    return {
      typeInstanceId: isSet(object.typeInstanceId)
        ? String(object.typeInstanceId)
        : "",
      newResourceVersion: isSet(object.newResourceVersion)
        ? Number(object.newResourceVersion)
        : 0,
      newValue: isSet(object.newValue)
        ? bytesFromBase64(object.newValue)
        : new Uint8Array(),
      context: isSet(object.context)
        ? bytesFromBase64(object.context)
        : undefined,
      ownerId: isSet(object.ownerId) ? String(object.ownerId) : undefined,
    };
  },

  toJSON(message: OnUpdateRequest): unknown {
    const obj: any = {};
    message.typeInstanceId !== undefined &&
      (obj.typeInstanceId = message.typeInstanceId);
    message.newResourceVersion !== undefined &&
      (obj.newResourceVersion = Math.round(message.newResourceVersion));
    message.newValue !== undefined &&
      (obj.newValue = base64FromBytes(
        message.newValue !== undefined ? message.newValue : new Uint8Array()
      ));
    message.context !== undefined &&
      (obj.context =
        message.context !== undefined
          ? base64FromBytes(message.context)
          : undefined);
    message.ownerId !== undefined && (obj.ownerId = message.ownerId);
    return obj;
  },

  fromPartial(object: DeepPartial<OnUpdateRequest>): OnUpdateRequest {
    const message = createBaseOnUpdateRequest();
    message.typeInstanceId = object.typeInstanceId ?? "";
    message.newResourceVersion = object.newResourceVersion ?? 0;
    message.newValue = object.newValue ?? new Uint8Array();
    message.context = object.context ?? undefined;
    message.ownerId = object.ownerId ?? undefined;
    return message;
  },
};

function createBaseOnUpdateResponse(): OnUpdateResponse {
  return { context: undefined };
}

export const OnUpdateResponse = {
  encode(
    message: OnUpdateResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.context !== undefined) {
      writer.uint32(10).bytes(message.context);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OnUpdateResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseOnUpdateResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): OnUpdateResponse {
    return {
      context: isSet(object.context)
        ? bytesFromBase64(object.context)
        : undefined,
    };
  },

  toJSON(message: OnUpdateResponse): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context =
        message.context !== undefined
          ? base64FromBytes(message.context)
          : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<OnUpdateResponse>): OnUpdateResponse {
    const message = createBaseOnUpdateResponse();
    message.context = object.context ?? undefined;
    return message;
  },
};

function createBaseOnDeleteRequest(): OnDeleteRequest {
  return { typeInstanceId: "", context: undefined, ownerId: undefined };
}

export const OnDeleteRequest = {
  encode(
    message: OnDeleteRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.typeInstanceId !== "") {
      writer.uint32(10).string(message.typeInstanceId);
    }
    if (message.context !== undefined) {
      writer.uint32(18).bytes(message.context);
    }
    if (message.ownerId !== undefined) {
      writer.uint32(26).string(message.ownerId);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OnDeleteRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseOnDeleteRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.typeInstanceId = reader.string();
          break;
        case 2:
          message.context = reader.bytes();
          break;
        case 3:
          message.ownerId = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): OnDeleteRequest {
    return {
      typeInstanceId: isSet(object.typeInstanceId)
        ? String(object.typeInstanceId)
        : "",
      context: isSet(object.context)
        ? bytesFromBase64(object.context)
        : undefined,
      ownerId: isSet(object.ownerId) ? String(object.ownerId) : undefined,
    };
  },

  toJSON(message: OnDeleteRequest): unknown {
    const obj: any = {};
    message.typeInstanceId !== undefined &&
      (obj.typeInstanceId = message.typeInstanceId);
    message.context !== undefined &&
      (obj.context =
        message.context !== undefined
          ? base64FromBytes(message.context)
          : undefined);
    message.ownerId !== undefined && (obj.ownerId = message.ownerId);
    return obj;
  },

  fromPartial(object: DeepPartial<OnDeleteRequest>): OnDeleteRequest {
    const message = createBaseOnDeleteRequest();
    message.typeInstanceId = object.typeInstanceId ?? "";
    message.context = object.context ?? undefined;
    message.ownerId = object.ownerId ?? undefined;
    return message;
  },
};

function createBaseOnDeleteResponse(): OnDeleteResponse {
  return {};
}

export const OnDeleteResponse = {
  encode(
    _: OnDeleteResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OnDeleteResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseOnDeleteResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): OnDeleteResponse {
    return {};
  },

  toJSON(_: OnDeleteResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<OnDeleteResponse>): OnDeleteResponse {
    const message = createBaseOnDeleteResponse();
    return message;
  },
};

function createBaseGetValueRequest(): GetValueRequest {
  return { typeInstanceId: "", resourceVersion: 0, context: new Uint8Array() };
}

export const GetValueRequest = {
  encode(
    message: GetValueRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.typeInstanceId !== "") {
      writer.uint32(10).string(message.typeInstanceId);
    }
    if (message.resourceVersion !== 0) {
      writer.uint32(16).uint32(message.resourceVersion);
    }
    if (message.context.length !== 0) {
      writer.uint32(26).bytes(message.context);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetValueRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseGetValueRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.typeInstanceId = reader.string();
          break;
        case 2:
          message.resourceVersion = reader.uint32();
          break;
        case 3:
          message.context = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetValueRequest {
    return {
      typeInstanceId: isSet(object.typeInstanceId)
        ? String(object.typeInstanceId)
        : "",
      resourceVersion: isSet(object.resourceVersion)
        ? Number(object.resourceVersion)
        : 0,
      context: isSet(object.context)
        ? bytesFromBase64(object.context)
        : new Uint8Array(),
    };
  },

  toJSON(message: GetValueRequest): unknown {
    const obj: any = {};
    message.typeInstanceId !== undefined &&
      (obj.typeInstanceId = message.typeInstanceId);
    message.resourceVersion !== undefined &&
      (obj.resourceVersion = Math.round(message.resourceVersion));
    message.context !== undefined &&
      (obj.context = base64FromBytes(
        message.context !== undefined ? message.context : new Uint8Array()
      ));
    return obj;
  },

  fromPartial(object: DeepPartial<GetValueRequest>): GetValueRequest {
    const message = createBaseGetValueRequest();
    message.typeInstanceId = object.typeInstanceId ?? "";
    message.resourceVersion = object.resourceVersion ?? 0;
    message.context = object.context ?? new Uint8Array();
    return message;
  },
};

function createBaseGetValueResponse(): GetValueResponse {
  return { value: undefined };
}

export const GetValueResponse = {
  encode(
    message: GetValueResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.value !== undefined) {
      writer.uint32(10).bytes(message.value);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetValueResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseGetValueResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.value = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetValueResponse {
    return {
      value: isSet(object.value) ? bytesFromBase64(object.value) : undefined,
    };
  },

  toJSON(message: GetValueResponse): unknown {
    const obj: any = {};
    message.value !== undefined &&
      (obj.value =
        message.value !== undefined
          ? base64FromBytes(message.value)
          : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<GetValueResponse>): GetValueResponse {
    const message = createBaseGetValueResponse();
    message.value = object.value ?? undefined;
    return message;
  },
};

function createBaseGetLockedByRequest(): GetLockedByRequest {
  return { typeInstanceId: "", context: new Uint8Array() };
}

export const GetLockedByRequest = {
  encode(
    message: GetLockedByRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.typeInstanceId !== "") {
      writer.uint32(10).string(message.typeInstanceId);
    }
    if (message.context.length !== 0) {
      writer.uint32(18).bytes(message.context);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetLockedByRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseGetLockedByRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.typeInstanceId = reader.string();
          break;
        case 2:
          message.context = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetLockedByRequest {
    return {
      typeInstanceId: isSet(object.typeInstanceId)
        ? String(object.typeInstanceId)
        : "",
      context: isSet(object.context)
        ? bytesFromBase64(object.context)
        : new Uint8Array(),
    };
  },

  toJSON(message: GetLockedByRequest): unknown {
    const obj: any = {};
    message.typeInstanceId !== undefined &&
      (obj.typeInstanceId = message.typeInstanceId);
    message.context !== undefined &&
      (obj.context = base64FromBytes(
        message.context !== undefined ? message.context : new Uint8Array()
      ));
    return obj;
  },

  fromPartial(object: DeepPartial<GetLockedByRequest>): GetLockedByRequest {
    const message = createBaseGetLockedByRequest();
    message.typeInstanceId = object.typeInstanceId ?? "";
    message.context = object.context ?? new Uint8Array();
    return message;
  },
};

function createBaseGetLockedByResponse(): GetLockedByResponse {
  return { lockedBy: undefined };
}

export const GetLockedByResponse = {
  encode(
    message: GetLockedByResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.lockedBy !== undefined) {
      writer.uint32(10).string(message.lockedBy);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetLockedByResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseGetLockedByResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.lockedBy = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetLockedByResponse {
    return {
      lockedBy: isSet(object.lockedBy) ? String(object.lockedBy) : undefined,
    };
  },

  toJSON(message: GetLockedByResponse): unknown {
    const obj: any = {};
    message.lockedBy !== undefined && (obj.lockedBy = message.lockedBy);
    return obj;
  },

  fromPartial(object: DeepPartial<GetLockedByResponse>): GetLockedByResponse {
    const message = createBaseGetLockedByResponse();
    message.lockedBy = object.lockedBy ?? undefined;
    return message;
  },
};

function createBaseOnLockRequest(): OnLockRequest {
  return { typeInstanceId: "", context: new Uint8Array(), lockedBy: "" };
}

export const OnLockRequest = {
  encode(
    message: OnLockRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.typeInstanceId !== "") {
      writer.uint32(10).string(message.typeInstanceId);
    }
    if (message.context.length !== 0) {
      writer.uint32(18).bytes(message.context);
    }
    if (message.lockedBy !== "") {
      writer.uint32(26).string(message.lockedBy);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OnLockRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseOnLockRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.typeInstanceId = reader.string();
          break;
        case 2:
          message.context = reader.bytes();
          break;
        case 3:
          message.lockedBy = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): OnLockRequest {
    return {
      typeInstanceId: isSet(object.typeInstanceId)
        ? String(object.typeInstanceId)
        : "",
      context: isSet(object.context)
        ? bytesFromBase64(object.context)
        : new Uint8Array(),
      lockedBy: isSet(object.lockedBy) ? String(object.lockedBy) : "",
    };
  },

  toJSON(message: OnLockRequest): unknown {
    const obj: any = {};
    message.typeInstanceId !== undefined &&
      (obj.typeInstanceId = message.typeInstanceId);
    message.context !== undefined &&
      (obj.context = base64FromBytes(
        message.context !== undefined ? message.context : new Uint8Array()
      ));
    message.lockedBy !== undefined && (obj.lockedBy = message.lockedBy);
    return obj;
  },

  fromPartial(object: DeepPartial<OnLockRequest>): OnLockRequest {
    const message = createBaseOnLockRequest();
    message.typeInstanceId = object.typeInstanceId ?? "";
    message.context = object.context ?? new Uint8Array();
    message.lockedBy = object.lockedBy ?? "";
    return message;
  },
};

function createBaseOnLockResponse(): OnLockResponse {
  return {};
}

export const OnLockResponse = {
  encode(
    _: OnLockResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OnLockResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseOnLockResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): OnLockResponse {
    return {};
  },

  toJSON(_: OnLockResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<OnLockResponse>): OnLockResponse {
    const message = createBaseOnLockResponse();
    return message;
  },
};

function createBaseOnUnlockRequest(): OnUnlockRequest {
  return { typeInstanceId: "", context: new Uint8Array() };
}

export const OnUnlockRequest = {
  encode(
    message: OnUnlockRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.typeInstanceId !== "") {
      writer.uint32(10).string(message.typeInstanceId);
    }
    if (message.context.length !== 0) {
      writer.uint32(18).bytes(message.context);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OnUnlockRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseOnUnlockRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.typeInstanceId = reader.string();
          break;
        case 2:
          message.context = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): OnUnlockRequest {
    return {
      typeInstanceId: isSet(object.typeInstanceId)
        ? String(object.typeInstanceId)
        : "",
      context: isSet(object.context)
        ? bytesFromBase64(object.context)
        : new Uint8Array(),
    };
  },

  toJSON(message: OnUnlockRequest): unknown {
    const obj: any = {};
    message.typeInstanceId !== undefined &&
      (obj.typeInstanceId = message.typeInstanceId);
    message.context !== undefined &&
      (obj.context = base64FromBytes(
        message.context !== undefined ? message.context : new Uint8Array()
      ));
    return obj;
  },

  fromPartial(object: DeepPartial<OnUnlockRequest>): OnUnlockRequest {
    const message = createBaseOnUnlockRequest();
    message.typeInstanceId = object.typeInstanceId ?? "";
    message.context = object.context ?? new Uint8Array();
    return message;
  },
};

function createBaseOnUnlockResponse(): OnUnlockResponse {
  return {};
}

export const OnUnlockResponse = {
  encode(
    _: OnUnlockResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OnUnlockResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseOnUnlockResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): OnUnlockResponse {
    return {};
  },

  toJSON(_: OnUnlockResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<OnUnlockResponse>): OnUnlockResponse {
    const message = createBaseOnUnlockResponse();
    return message;
  },
};

export const StorageBackendDefinition = {
  name: "StorageBackend",
  fullName: "storage_backend.StorageBackend",
  methods: {
    /** value */
    getValue: {
      name: "GetValue",
      requestType: GetValueRequest,
      requestStream: false,
      responseType: GetValueResponse,
      responseStream: false,
      options: {},
    },
    onCreate: {
      name: "OnCreate",
      requestType: OnCreateRequest,
      requestStream: false,
      responseType: OnCreateResponse,
      responseStream: false,
      options: {},
    },
    onUpdate: {
      name: "OnUpdate",
      requestType: OnUpdateRequest,
      requestStream: false,
      responseType: OnUpdateResponse,
      responseStream: false,
      options: {},
    },
    onDelete: {
      name: "OnDelete",
      requestType: OnDeleteRequest,
      requestStream: false,
      responseType: OnDeleteResponse,
      responseStream: false,
      options: {},
    },
    /** lock */
    getLockedBy: {
      name: "GetLockedBy",
      requestType: GetLockedByRequest,
      requestStream: false,
      responseType: GetLockedByResponse,
      responseStream: false,
      options: {},
    },
    onLock: {
      name: "OnLock",
      requestType: OnLockRequest,
      requestStream: false,
      responseType: OnLockResponse,
      responseStream: false,
      options: {},
    },
    onUnlock: {
      name: "OnUnlock",
      requestType: OnUnlockRequest,
      requestStream: false,
      responseType: OnUnlockResponse,
      responseStream: false,
      options: {},
    },
  },
} as const;

declare var self: any | undefined;
declare var window: any | undefined;
declare var global: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

const atob: (b64: string) => string =
  globalThis.atob ||
  ((b64) => globalThis.Buffer.from(b64, "base64").toString("binary"));
function bytesFromBase64(b64: string): Uint8Array {
  const bin = atob(b64);
  const arr = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; ++i) {
    arr[i] = bin.charCodeAt(i);
  }
  return arr;
}

const btoa: (bin: string) => string =
  globalThis.btoa ||
  ((bin) => globalThis.Buffer.from(bin, "binary").toString("base64"));
function base64FromBytes(arr: Uint8Array): string {
  const bin: string[] = [];
  for (const byte of arr) {
    bin.push(String.fromCharCode(byte));
  }
  return btoa(bin.join(""));
}

type Builtin =
  | Date
  | Function
  | Uint8Array
  | string
  | number
  | boolean
  | undefined;

export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any;
  _m0.configure();
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}
