import type * as grpc from '@grpc/grpc-js';
import type { MessageTypeDefinition } from '@grpc/proto-loader';

import type { StorageBackendClient as _storage_backend_StorageBackendClient, StorageBackendDefinition as _storage_backend_StorageBackendDefinition } from './storage_backend/StorageBackend';

type SubtypeConstructor<Constructor extends new (...args: any) => any, Subtype> = {
  new(...args: ConstructorParameters<Constructor>): Subtype;
};

export interface ProtoGrpcType {
  storage_backend: {
    GetLockedByRequest: MessageTypeDefinition
    GetLockedByResponse: MessageTypeDefinition
    GetValueRequest: MessageTypeDefinition
    GetValueResponse: MessageTypeDefinition
    OnCreateRequest: MessageTypeDefinition
    OnCreateResponse: MessageTypeDefinition
    OnDeleteRequest: MessageTypeDefinition
    OnDeleteResponse: MessageTypeDefinition
    OnLockRequest: MessageTypeDefinition
    OnLockResponse: MessageTypeDefinition
    OnUnlockRequest: MessageTypeDefinition
    OnUnlockResponse: MessageTypeDefinition
    OnUpdateRequest: MessageTypeDefinition
    OnUpdateResponse: MessageTypeDefinition
    StorageBackend: SubtypeConstructor<typeof grpc.Client, _storage_backend_StorageBackendClient> & { service: _storage_backend_StorageBackendDefinition }
    TypeInstanceResourceVersion: MessageTypeDefinition
  }
}

