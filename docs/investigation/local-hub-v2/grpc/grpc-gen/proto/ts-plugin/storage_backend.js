"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.StorageBackendDefinition = exports.OnUnlockResponse = exports.OnUnlockRequest = exports.OnLockResponse = exports.OnLockRequest = exports.GetLockedByResponse = exports.GetLockedByRequest = exports.GetValueResponse = exports.GetValueRequest = exports.OnDeleteResponse = exports.OnDeleteRequest = exports.OnUpdateResponse = exports.OnUpdateRequest = exports.TypeInstanceResourceVersion = exports.OnCreateResponse = exports.OnCreateRequest = exports.protobufPackage = void 0;
/* eslint-disable */
const long_1 = __importDefault(require("long"));
const minimal_1 = __importDefault(require("protobufjs/minimal"));
exports.protobufPackage = 'storage_backend';
function createBaseOnCreateRequest() {
    return {
        typeinstanceId: '',
        value: new Uint8Array(),
        context: new Uint8Array(),
    };
}
exports.OnCreateRequest = {
    encode(message, writer = minimal_1.default.Writer.create()) {
        if (message.typeinstanceId !== '') {
            writer.uint32(10).string(message.typeinstanceId);
        }
        if (message.value.length !== 0) {
            writer.uint32(18).bytes(message.value);
        }
        if (message.context.length !== 0) {
            writer.uint32(26).bytes(message.context);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseOnCreateRequest();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.typeinstanceId = reader.string();
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
    fromJSON(object) {
        return {
            typeinstanceId: isSet(object.typeinstanceId)
                ? String(object.typeinstanceId)
                : '',
            value: isSet(object.value)
                ? bytesFromBase64(object.value)
                : new Uint8Array(),
            context: isSet(object.context)
                ? bytesFromBase64(object.context)
                : new Uint8Array(),
        };
    },
    toJSON(message) {
        const obj = {};
        message.typeinstanceId !== undefined &&
            (obj.typeinstanceId = message.typeinstanceId);
        message.value !== undefined &&
            (obj.value = base64FromBytes(message.value !== undefined ? message.value : new Uint8Array()));
        message.context !== undefined &&
            (obj.context = base64FromBytes(message.context !== undefined ? message.context : new Uint8Array()));
        return obj;
    },
    fromPartial(object) {
        var _a, _b, _c;
        const message = createBaseOnCreateRequest();
        message.typeinstanceId = (_a = object.typeinstanceId) !== null && _a !== void 0 ? _a : '';
        message.value = (_b = object.value) !== null && _b !== void 0 ? _b : new Uint8Array();
        message.context = (_c = object.context) !== null && _c !== void 0 ? _c : new Uint8Array();
        return message;
    },
};
function createBaseOnCreateResponse() {
    return { context: undefined };
}
exports.OnCreateResponse = {
    encode(message, writer = minimal_1.default.Writer.create()) {
        if (message.context !== undefined) {
            writer.uint32(10).bytes(message.context);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
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
    fromJSON(object) {
        return {
            context: isSet(object.context)
                ? bytesFromBase64(object.context)
                : undefined,
        };
    },
    toJSON(message) {
        const obj = {};
        message.context !== undefined &&
            (obj.context =
                message.context !== undefined
                    ? base64FromBytes(message.context)
                    : undefined);
        return obj;
    },
    fromPartial(object) {
        var _a;
        const message = createBaseOnCreateResponse();
        message.context = (_a = object.context) !== null && _a !== void 0 ? _a : undefined;
        return message;
    },
};
function createBaseTypeInstanceResourceVersion() {
    return { resourceVersion: 0, value: new Uint8Array() };
}
exports.TypeInstanceResourceVersion = {
    encode(message, writer = minimal_1.default.Writer.create()) {
        if (message.resourceVersion !== 0) {
            writer.uint32(8).uint32(message.resourceVersion);
        }
        if (message.value.length !== 0) {
            writer.uint32(18).bytes(message.value);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
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
    fromJSON(object) {
        return {
            resourceVersion: isSet(object.resourceVersion)
                ? Number(object.resourceVersion)
                : 0,
            value: isSet(object.value)
                ? bytesFromBase64(object.value)
                : new Uint8Array(),
        };
    },
    toJSON(message) {
        const obj = {};
        message.resourceVersion !== undefined &&
            (obj.resourceVersion = Math.round(message.resourceVersion));
        message.value !== undefined &&
            (obj.value = base64FromBytes(message.value !== undefined ? message.value : new Uint8Array()));
        return obj;
    },
    fromPartial(object) {
        var _a, _b;
        const message = createBaseTypeInstanceResourceVersion();
        message.resourceVersion = (_a = object.resourceVersion) !== null && _a !== void 0 ? _a : 0;
        message.value = (_b = object.value) !== null && _b !== void 0 ? _b : new Uint8Array();
        return message;
    },
};
function createBaseOnUpdateRequest() {
    return {
        typeinstanceId: '',
        newResourceVersion: 0,
        newValue: new Uint8Array(),
        context: undefined,
    };
}
exports.OnUpdateRequest = {
    encode(message, writer = minimal_1.default.Writer.create()) {
        if (message.typeinstanceId !== '') {
            writer.uint32(10).string(message.typeinstanceId);
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
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseOnUpdateRequest();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.typeinstanceId = reader.string();
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
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        return {
            typeinstanceId: isSet(object.typeinstanceId)
                ? String(object.typeinstanceId)
                : '',
            newResourceVersion: isSet(object.newResourceVersion)
                ? Number(object.newResourceVersion)
                : 0,
            newValue: isSet(object.newValue)
                ? bytesFromBase64(object.newValue)
                : new Uint8Array(),
            context: isSet(object.context)
                ? bytesFromBase64(object.context)
                : undefined,
        };
    },
    toJSON(message) {
        const obj = {};
        message.typeinstanceId !== undefined &&
            (obj.typeinstanceId = message.typeinstanceId);
        message.newResourceVersion !== undefined &&
            (obj.newResourceVersion = Math.round(message.newResourceVersion));
        message.newValue !== undefined &&
            (obj.newValue = base64FromBytes(message.newValue !== undefined ? message.newValue : new Uint8Array()));
        message.context !== undefined &&
            (obj.context =
                message.context !== undefined
                    ? base64FromBytes(message.context)
                    : undefined);
        return obj;
    },
    fromPartial(object) {
        var _a, _b, _c, _d;
        const message = createBaseOnUpdateRequest();
        message.typeinstanceId = (_a = object.typeinstanceId) !== null && _a !== void 0 ? _a : '';
        message.newResourceVersion = (_b = object.newResourceVersion) !== null && _b !== void 0 ? _b : 0;
        message.newValue = (_c = object.newValue) !== null && _c !== void 0 ? _c : new Uint8Array();
        message.context = (_d = object.context) !== null && _d !== void 0 ? _d : undefined;
        return message;
    },
};
function createBaseOnUpdateResponse() {
    return { context: undefined };
}
exports.OnUpdateResponse = {
    encode(message, writer = minimal_1.default.Writer.create()) {
        if (message.context !== undefined) {
            writer.uint32(10).bytes(message.context);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
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
    fromJSON(object) {
        return {
            context: isSet(object.context)
                ? bytesFromBase64(object.context)
                : undefined,
        };
    },
    toJSON(message) {
        const obj = {};
        message.context !== undefined &&
            (obj.context =
                message.context !== undefined
                    ? base64FromBytes(message.context)
                    : undefined);
        return obj;
    },
    fromPartial(object) {
        var _a;
        const message = createBaseOnUpdateResponse();
        message.context = (_a = object.context) !== null && _a !== void 0 ? _a : undefined;
        return message;
    },
};
function createBaseOnDeleteRequest() {
    return { typeinstanceId: '', context: new Uint8Array() };
}
exports.OnDeleteRequest = {
    encode(message, writer = minimal_1.default.Writer.create()) {
        if (message.typeinstanceId !== '') {
            writer.uint32(10).string(message.typeinstanceId);
        }
        if (message.context.length !== 0) {
            writer.uint32(18).bytes(message.context);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseOnDeleteRequest();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.typeinstanceId = reader.string();
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
    fromJSON(object) {
        return {
            typeinstanceId: isSet(object.typeinstanceId)
                ? String(object.typeinstanceId)
                : '',
            context: isSet(object.context)
                ? bytesFromBase64(object.context)
                : new Uint8Array(),
        };
    },
    toJSON(message) {
        const obj = {};
        message.typeinstanceId !== undefined &&
            (obj.typeinstanceId = message.typeinstanceId);
        message.context !== undefined &&
            (obj.context = base64FromBytes(message.context !== undefined ? message.context : new Uint8Array()));
        return obj;
    },
    fromPartial(object) {
        var _a, _b;
        const message = createBaseOnDeleteRequest();
        message.typeinstanceId = (_a = object.typeinstanceId) !== null && _a !== void 0 ? _a : '';
        message.context = (_b = object.context) !== null && _b !== void 0 ? _b : new Uint8Array();
        return message;
    },
};
function createBaseOnDeleteResponse() {
    return {};
}
exports.OnDeleteResponse = {
    encode(_, writer = minimal_1.default.Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
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
    fromJSON(_) {
        return {};
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = createBaseOnDeleteResponse();
        return message;
    },
};
function createBaseGetValueRequest() {
    return { typeinstanceId: '', resourceVersion: 0, context: new Uint8Array() };
}
exports.GetValueRequest = {
    encode(message, writer = minimal_1.default.Writer.create()) {
        if (message.typeinstanceId !== '') {
            writer.uint32(10).string(message.typeinstanceId);
        }
        if (message.resourceVersion !== 0) {
            writer.uint32(16).uint32(message.resourceVersion);
        }
        if (message.context.length !== 0) {
            writer.uint32(26).bytes(message.context);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseGetValueRequest();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.typeinstanceId = reader.string();
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
    fromJSON(object) {
        return {
            typeinstanceId: isSet(object.typeinstanceId)
                ? String(object.typeinstanceId)
                : '',
            resourceVersion: isSet(object.resourceVersion)
                ? Number(object.resourceVersion)
                : 0,
            context: isSet(object.context)
                ? bytesFromBase64(object.context)
                : new Uint8Array(),
        };
    },
    toJSON(message) {
        const obj = {};
        message.typeinstanceId !== undefined &&
            (obj.typeinstanceId = message.typeinstanceId);
        message.resourceVersion !== undefined &&
            (obj.resourceVersion = Math.round(message.resourceVersion));
        message.context !== undefined &&
            (obj.context = base64FromBytes(message.context !== undefined ? message.context : new Uint8Array()));
        return obj;
    },
    fromPartial(object) {
        var _a, _b, _c;
        const message = createBaseGetValueRequest();
        message.typeinstanceId = (_a = object.typeinstanceId) !== null && _a !== void 0 ? _a : '';
        message.resourceVersion = (_b = object.resourceVersion) !== null && _b !== void 0 ? _b : 0;
        message.context = (_c = object.context) !== null && _c !== void 0 ? _c : new Uint8Array();
        return message;
    },
};
function createBaseGetValueResponse() {
    return { value: undefined };
}
exports.GetValueResponse = {
    encode(message, writer = minimal_1.default.Writer.create()) {
        if (message.value !== undefined) {
            writer.uint32(10).bytes(message.value);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
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
    fromJSON(object) {
        return {
            value: isSet(object.value) ? bytesFromBase64(object.value) : undefined,
        };
    },
    toJSON(message) {
        const obj = {};
        message.value !== undefined &&
            (obj.value =
                message.value !== undefined
                    ? base64FromBytes(message.value)
                    : undefined);
        return obj;
    },
    fromPartial(object) {
        var _a;
        const message = createBaseGetValueResponse();
        message.value = (_a = object.value) !== null && _a !== void 0 ? _a : undefined;
        return message;
    },
};
function createBaseGetLockedByRequest() {
    return { typeinstanceId: '', context: new Uint8Array() };
}
exports.GetLockedByRequest = {
    encode(message, writer = minimal_1.default.Writer.create()) {
        if (message.typeinstanceId !== '') {
            writer.uint32(10).string(message.typeinstanceId);
        }
        if (message.context.length !== 0) {
            writer.uint32(18).bytes(message.context);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseGetLockedByRequest();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.typeinstanceId = reader.string();
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
    fromJSON(object) {
        return {
            typeinstanceId: isSet(object.typeinstanceId)
                ? String(object.typeinstanceId)
                : '',
            context: isSet(object.context)
                ? bytesFromBase64(object.context)
                : new Uint8Array(),
        };
    },
    toJSON(message) {
        const obj = {};
        message.typeinstanceId !== undefined &&
            (obj.typeinstanceId = message.typeinstanceId);
        message.context !== undefined &&
            (obj.context = base64FromBytes(message.context !== undefined ? message.context : new Uint8Array()));
        return obj;
    },
    fromPartial(object) {
        var _a, _b;
        const message = createBaseGetLockedByRequest();
        message.typeinstanceId = (_a = object.typeinstanceId) !== null && _a !== void 0 ? _a : '';
        message.context = (_b = object.context) !== null && _b !== void 0 ? _b : new Uint8Array();
        return message;
    },
};
function createBaseGetLockedByResponse() {
    return { lockedBy: undefined };
}
exports.GetLockedByResponse = {
    encode(message, writer = minimal_1.default.Writer.create()) {
        if (message.lockedBy !== undefined) {
            writer.uint32(10).string(message.lockedBy);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
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
    fromJSON(object) {
        return {
            lockedBy: isSet(object.lockedBy) ? String(object.lockedBy) : undefined,
        };
    },
    toJSON(message) {
        const obj = {};
        message.lockedBy !== undefined && (obj.lockedBy = message.lockedBy);
        return obj;
    },
    fromPartial(object) {
        var _a;
        const message = createBaseGetLockedByResponse();
        message.lockedBy = (_a = object.lockedBy) !== null && _a !== void 0 ? _a : undefined;
        return message;
    },
};
function createBaseOnLockRequest() {
    return { typeinstanceId: '', context: new Uint8Array(), lockedBy: '' };
}
exports.OnLockRequest = {
    encode(message, writer = minimal_1.default.Writer.create()) {
        if (message.typeinstanceId !== '') {
            writer.uint32(10).string(message.typeinstanceId);
        }
        if (message.context.length !== 0) {
            writer.uint32(18).bytes(message.context);
        }
        if (message.lockedBy !== '') {
            writer.uint32(26).string(message.lockedBy);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseOnLockRequest();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.typeinstanceId = reader.string();
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
    fromJSON(object) {
        return {
            typeinstanceId: isSet(object.typeinstanceId)
                ? String(object.typeinstanceId)
                : '',
            context: isSet(object.context)
                ? bytesFromBase64(object.context)
                : new Uint8Array(),
            lockedBy: isSet(object.lockedBy) ? String(object.lockedBy) : '',
        };
    },
    toJSON(message) {
        const obj = {};
        message.typeinstanceId !== undefined &&
            (obj.typeinstanceId = message.typeinstanceId);
        message.context !== undefined &&
            (obj.context = base64FromBytes(message.context !== undefined ? message.context : new Uint8Array()));
        message.lockedBy !== undefined && (obj.lockedBy = message.lockedBy);
        return obj;
    },
    fromPartial(object) {
        var _a, _b, _c;
        const message = createBaseOnLockRequest();
        message.typeinstanceId = (_a = object.typeinstanceId) !== null && _a !== void 0 ? _a : '';
        message.context = (_b = object.context) !== null && _b !== void 0 ? _b : new Uint8Array();
        message.lockedBy = (_c = object.lockedBy) !== null && _c !== void 0 ? _c : '';
        return message;
    },
};
function createBaseOnLockResponse() {
    return {};
}
exports.OnLockResponse = {
    encode(_, writer = minimal_1.default.Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
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
    fromJSON(_) {
        return {};
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = createBaseOnLockResponse();
        return message;
    },
};
function createBaseOnUnlockRequest() {
    return { typeinstanceId: '', context: new Uint8Array() };
}
exports.OnUnlockRequest = {
    encode(message, writer = minimal_1.default.Writer.create()) {
        if (message.typeinstanceId !== '') {
            writer.uint32(10).string(message.typeinstanceId);
        }
        if (message.context.length !== 0) {
            writer.uint32(18).bytes(message.context);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseOnUnlockRequest();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.typeinstanceId = reader.string();
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
    fromJSON(object) {
        return {
            typeinstanceId: isSet(object.typeinstanceId)
                ? String(object.typeinstanceId)
                : '',
            context: isSet(object.context)
                ? bytesFromBase64(object.context)
                : new Uint8Array(),
        };
    },
    toJSON(message) {
        const obj = {};
        message.typeinstanceId !== undefined &&
            (obj.typeinstanceId = message.typeinstanceId);
        message.context !== undefined &&
            (obj.context = base64FromBytes(message.context !== undefined ? message.context : new Uint8Array()));
        return obj;
    },
    fromPartial(object) {
        var _a, _b;
        const message = createBaseOnUnlockRequest();
        message.typeinstanceId = (_a = object.typeinstanceId) !== null && _a !== void 0 ? _a : '';
        message.context = (_b = object.context) !== null && _b !== void 0 ? _b : new Uint8Array();
        return message;
    },
};
function createBaseOnUnlockResponse() {
    return {};
}
exports.OnUnlockResponse = {
    encode(_, writer = minimal_1.default.Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof minimal_1.default.Reader ? input : new minimal_1.default.Reader(input);
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
    fromJSON(_) {
        return {};
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = createBaseOnUnlockResponse();
        return message;
    },
};
exports.StorageBackendDefinition = {
    name: 'StorageBackend',
    fullName: 'storage_backend.StorageBackend',
    methods: {
        /** value */
        getValue: {
            name: 'GetValue',
            requestType: exports.GetValueRequest,
            requestStream: false,
            responseType: exports.GetValueResponse,
            responseStream: false,
            options: {},
        },
        onCreate: {
            name: 'OnCreate',
            requestType: exports.OnCreateRequest,
            requestStream: false,
            responseType: exports.OnCreateResponse,
            responseStream: false,
            options: {},
        },
        onUpdate: {
            name: 'OnUpdate',
            requestType: exports.OnUpdateRequest,
            requestStream: false,
            responseType: exports.OnUpdateResponse,
            responseStream: false,
            options: {},
        },
        onDelete: {
            name: 'OnDelete',
            requestType: exports.OnDeleteRequest,
            requestStream: false,
            responseType: exports.OnDeleteResponse,
            responseStream: false,
            options: {},
        },
        /** lock */
        getLockedBy: {
            name: 'GetLockedBy',
            requestType: exports.GetLockedByRequest,
            requestStream: false,
            responseType: exports.GetLockedByResponse,
            responseStream: false,
            options: {},
        },
        onLock: {
            name: 'OnLock',
            requestType: exports.OnLockRequest,
            requestStream: false,
            responseType: exports.OnLockResponse,
            responseStream: false,
            options: {},
        },
        onUnlock: {
            name: 'OnUnlock',
            requestType: exports.OnUnlockRequest,
            requestStream: false,
            responseType: exports.OnUnlockResponse,
            responseStream: false,
            options: {},
        },
    },
};
var globalThis = (() => {
    if (typeof globalThis !== 'undefined')
        return globalThis;
    if (typeof self !== 'undefined')
        return self;
    if (typeof window !== 'undefined')
        return window;
    if (typeof global !== 'undefined')
        return global;
    throw 'Unable to locate global object';
})();
const atob = globalThis.atob ||
    ((b64) => globalThis.Buffer.from(b64, 'base64').toString('binary'));
function bytesFromBase64(b64) {
    const bin = atob(b64);
    const arr = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; ++i) {
        arr[i] = bin.charCodeAt(i);
    }
    return arr;
}
const btoa = globalThis.btoa ||
    ((bin) => globalThis.Buffer.from(bin, 'binary').toString('base64'));
function base64FromBytes(arr) {
    const bin = [];
    for (const byte of arr) {
        bin.push(String.fromCharCode(byte));
    }
    return btoa(bin.join(''));
}
if (minimal_1.default.util.Long !== long_1.default) {
    minimal_1.default.util.Long = long_1.default;
    minimal_1.default.configure();
}
function isSet(value) {
    return value !== null && value !== undefined;
}
