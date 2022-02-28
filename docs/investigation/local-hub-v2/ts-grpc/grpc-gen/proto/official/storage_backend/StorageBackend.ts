// Original file: ../../../../../hub-js/proto/storage_backend.proto

import type * as grpc from '@grpc/grpc-js'
import type { MethodDefinition } from '@grpc/proto-loader'
import type { GetLockedByRequest as _storage_backend_GetLockedByRequest, GetLockedByRequest__Output as _storage_backend_GetLockedByRequest__Output } from '../storage_backend/GetLockedByRequest';
import type { GetLockedByResponse as _storage_backend_GetLockedByResponse, GetLockedByResponse__Output as _storage_backend_GetLockedByResponse__Output } from '../storage_backend/GetLockedByResponse';
import type { GetValueRequest as _storage_backend_GetValueRequest, GetValueRequest__Output as _storage_backend_GetValueRequest__Output } from '../storage_backend/GetValueRequest';
import type { GetValueResponse as _storage_backend_GetValueResponse, GetValueResponse__Output as _storage_backend_GetValueResponse__Output } from '../storage_backend/GetValueResponse';
import type { OnCreateRequest as _storage_backend_OnCreateRequest, OnCreateRequest__Output as _storage_backend_OnCreateRequest__Output } from '../storage_backend/OnCreateRequest';
import type { OnCreateResponse as _storage_backend_OnCreateResponse, OnCreateResponse__Output as _storage_backend_OnCreateResponse__Output } from '../storage_backend/OnCreateResponse';
import type { OnDeleteRequest as _storage_backend_OnDeleteRequest, OnDeleteRequest__Output as _storage_backend_OnDeleteRequest__Output } from '../storage_backend/OnDeleteRequest';
import type { OnDeleteResponse as _storage_backend_OnDeleteResponse, OnDeleteResponse__Output as _storage_backend_OnDeleteResponse__Output } from '../storage_backend/OnDeleteResponse';
import type { OnLockRequest as _storage_backend_OnLockRequest, OnLockRequest__Output as _storage_backend_OnLockRequest__Output } from '../storage_backend/OnLockRequest';
import type { OnLockResponse as _storage_backend_OnLockResponse, OnLockResponse__Output as _storage_backend_OnLockResponse__Output } from '../storage_backend/OnLockResponse';
import type { OnUnlockRequest as _storage_backend_OnUnlockRequest, OnUnlockRequest__Output as _storage_backend_OnUnlockRequest__Output } from '../storage_backend/OnUnlockRequest';
import type { OnUnlockResponse as _storage_backend_OnUnlockResponse, OnUnlockResponse__Output as _storage_backend_OnUnlockResponse__Output } from '../storage_backend/OnUnlockResponse';
import type { OnUpdateRequest as _storage_backend_OnUpdateRequest, OnUpdateRequest__Output as _storage_backend_OnUpdateRequest__Output } from '../storage_backend/OnUpdateRequest';
import type { OnUpdateResponse as _storage_backend_OnUpdateResponse, OnUpdateResponse__Output as _storage_backend_OnUpdateResponse__Output } from '../storage_backend/OnUpdateResponse';

export interface StorageBackendClient extends grpc.Client {
  GetLockedBy(argument: _storage_backend_GetLockedByRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_GetLockedByResponse__Output>): grpc.ClientUnaryCall;
  GetLockedBy(argument: _storage_backend_GetLockedByRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_GetLockedByResponse__Output>): grpc.ClientUnaryCall;
  GetLockedBy(argument: _storage_backend_GetLockedByRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_GetLockedByResponse__Output>): grpc.ClientUnaryCall;
  GetLockedBy(argument: _storage_backend_GetLockedByRequest, callback: grpc.requestCallback<_storage_backend_GetLockedByResponse__Output>): grpc.ClientUnaryCall;
  getLockedBy(argument: _storage_backend_GetLockedByRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_GetLockedByResponse__Output>): grpc.ClientUnaryCall;
  getLockedBy(argument: _storage_backend_GetLockedByRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_GetLockedByResponse__Output>): grpc.ClientUnaryCall;
  getLockedBy(argument: _storage_backend_GetLockedByRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_GetLockedByResponse__Output>): grpc.ClientUnaryCall;
  getLockedBy(argument: _storage_backend_GetLockedByRequest, callback: grpc.requestCallback<_storage_backend_GetLockedByResponse__Output>): grpc.ClientUnaryCall;
  
  GetValue(argument: _storage_backend_GetValueRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_GetValueResponse__Output>): grpc.ClientUnaryCall;
  GetValue(argument: _storage_backend_GetValueRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_GetValueResponse__Output>): grpc.ClientUnaryCall;
  GetValue(argument: _storage_backend_GetValueRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_GetValueResponse__Output>): grpc.ClientUnaryCall;
  GetValue(argument: _storage_backend_GetValueRequest, callback: grpc.requestCallback<_storage_backend_GetValueResponse__Output>): grpc.ClientUnaryCall;
  getValue(argument: _storage_backend_GetValueRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_GetValueResponse__Output>): grpc.ClientUnaryCall;
  getValue(argument: _storage_backend_GetValueRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_GetValueResponse__Output>): grpc.ClientUnaryCall;
  getValue(argument: _storage_backend_GetValueRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_GetValueResponse__Output>): grpc.ClientUnaryCall;
  getValue(argument: _storage_backend_GetValueRequest, callback: grpc.requestCallback<_storage_backend_GetValueResponse__Output>): grpc.ClientUnaryCall;
  
  OnCreate(argument: _storage_backend_OnCreateRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnCreateResponse__Output>): grpc.ClientUnaryCall;
  OnCreate(argument: _storage_backend_OnCreateRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_OnCreateResponse__Output>): grpc.ClientUnaryCall;
  OnCreate(argument: _storage_backend_OnCreateRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnCreateResponse__Output>): grpc.ClientUnaryCall;
  OnCreate(argument: _storage_backend_OnCreateRequest, callback: grpc.requestCallback<_storage_backend_OnCreateResponse__Output>): grpc.ClientUnaryCall;
  onCreate(argument: _storage_backend_OnCreateRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnCreateResponse__Output>): grpc.ClientUnaryCall;
  onCreate(argument: _storage_backend_OnCreateRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_OnCreateResponse__Output>): grpc.ClientUnaryCall;
  onCreate(argument: _storage_backend_OnCreateRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnCreateResponse__Output>): grpc.ClientUnaryCall;
  onCreate(argument: _storage_backend_OnCreateRequest, callback: grpc.requestCallback<_storage_backend_OnCreateResponse__Output>): grpc.ClientUnaryCall;
  
  OnDelete(argument: _storage_backend_OnDeleteRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnDeleteResponse__Output>): grpc.ClientUnaryCall;
  OnDelete(argument: _storage_backend_OnDeleteRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_OnDeleteResponse__Output>): grpc.ClientUnaryCall;
  OnDelete(argument: _storage_backend_OnDeleteRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnDeleteResponse__Output>): grpc.ClientUnaryCall;
  OnDelete(argument: _storage_backend_OnDeleteRequest, callback: grpc.requestCallback<_storage_backend_OnDeleteResponse__Output>): grpc.ClientUnaryCall;
  onDelete(argument: _storage_backend_OnDeleteRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnDeleteResponse__Output>): grpc.ClientUnaryCall;
  onDelete(argument: _storage_backend_OnDeleteRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_OnDeleteResponse__Output>): grpc.ClientUnaryCall;
  onDelete(argument: _storage_backend_OnDeleteRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnDeleteResponse__Output>): grpc.ClientUnaryCall;
  onDelete(argument: _storage_backend_OnDeleteRequest, callback: grpc.requestCallback<_storage_backend_OnDeleteResponse__Output>): grpc.ClientUnaryCall;
  
  OnLock(argument: _storage_backend_OnLockRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnLockResponse__Output>): grpc.ClientUnaryCall;
  OnLock(argument: _storage_backend_OnLockRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_OnLockResponse__Output>): grpc.ClientUnaryCall;
  OnLock(argument: _storage_backend_OnLockRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnLockResponse__Output>): grpc.ClientUnaryCall;
  OnLock(argument: _storage_backend_OnLockRequest, callback: grpc.requestCallback<_storage_backend_OnLockResponse__Output>): grpc.ClientUnaryCall;
  onLock(argument: _storage_backend_OnLockRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnLockResponse__Output>): grpc.ClientUnaryCall;
  onLock(argument: _storage_backend_OnLockRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_OnLockResponse__Output>): grpc.ClientUnaryCall;
  onLock(argument: _storage_backend_OnLockRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnLockResponse__Output>): grpc.ClientUnaryCall;
  onLock(argument: _storage_backend_OnLockRequest, callback: grpc.requestCallback<_storage_backend_OnLockResponse__Output>): grpc.ClientUnaryCall;
  
  OnUnlock(argument: _storage_backend_OnUnlockRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnUnlockResponse__Output>): grpc.ClientUnaryCall;
  OnUnlock(argument: _storage_backend_OnUnlockRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_OnUnlockResponse__Output>): grpc.ClientUnaryCall;
  OnUnlock(argument: _storage_backend_OnUnlockRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnUnlockResponse__Output>): grpc.ClientUnaryCall;
  OnUnlock(argument: _storage_backend_OnUnlockRequest, callback: grpc.requestCallback<_storage_backend_OnUnlockResponse__Output>): grpc.ClientUnaryCall;
  onUnlock(argument: _storage_backend_OnUnlockRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnUnlockResponse__Output>): grpc.ClientUnaryCall;
  onUnlock(argument: _storage_backend_OnUnlockRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_OnUnlockResponse__Output>): grpc.ClientUnaryCall;
  onUnlock(argument: _storage_backend_OnUnlockRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnUnlockResponse__Output>): grpc.ClientUnaryCall;
  onUnlock(argument: _storage_backend_OnUnlockRequest, callback: grpc.requestCallback<_storage_backend_OnUnlockResponse__Output>): grpc.ClientUnaryCall;
  
  OnUpdate(argument: _storage_backend_OnUpdateRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnUpdateResponse__Output>): grpc.ClientUnaryCall;
  OnUpdate(argument: _storage_backend_OnUpdateRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_OnUpdateResponse__Output>): grpc.ClientUnaryCall;
  OnUpdate(argument: _storage_backend_OnUpdateRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnUpdateResponse__Output>): grpc.ClientUnaryCall;
  OnUpdate(argument: _storage_backend_OnUpdateRequest, callback: grpc.requestCallback<_storage_backend_OnUpdateResponse__Output>): grpc.ClientUnaryCall;
  onUpdate(argument: _storage_backend_OnUpdateRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnUpdateResponse__Output>): grpc.ClientUnaryCall;
  onUpdate(argument: _storage_backend_OnUpdateRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_storage_backend_OnUpdateResponse__Output>): grpc.ClientUnaryCall;
  onUpdate(argument: _storage_backend_OnUpdateRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_storage_backend_OnUpdateResponse__Output>): grpc.ClientUnaryCall;
  onUpdate(argument: _storage_backend_OnUpdateRequest, callback: grpc.requestCallback<_storage_backend_OnUpdateResponse__Output>): grpc.ClientUnaryCall;
  
}

export interface StorageBackendHandlers extends grpc.UntypedServiceImplementation {
  GetLockedBy: grpc.handleUnaryCall<_storage_backend_GetLockedByRequest__Output, _storage_backend_GetLockedByResponse>;
  
  GetValue: grpc.handleUnaryCall<_storage_backend_GetValueRequest__Output, _storage_backend_GetValueResponse>;
  
  OnCreate: grpc.handleUnaryCall<_storage_backend_OnCreateRequest__Output, _storage_backend_OnCreateResponse>;
  
  OnDelete: grpc.handleUnaryCall<_storage_backend_OnDeleteRequest__Output, _storage_backend_OnDeleteResponse>;
  
  OnLock: grpc.handleUnaryCall<_storage_backend_OnLockRequest__Output, _storage_backend_OnLockResponse>;
  
  OnUnlock: grpc.handleUnaryCall<_storage_backend_OnUnlockRequest__Output, _storage_backend_OnUnlockResponse>;
  
  OnUpdate: grpc.handleUnaryCall<_storage_backend_OnUpdateRequest__Output, _storage_backend_OnUpdateResponse>;
  
}

export interface StorageBackendDefinition extends grpc.ServiceDefinition {
  GetLockedBy: MethodDefinition<_storage_backend_GetLockedByRequest, _storage_backend_GetLockedByResponse, _storage_backend_GetLockedByRequest__Output, _storage_backend_GetLockedByResponse__Output>
  GetValue: MethodDefinition<_storage_backend_GetValueRequest, _storage_backend_GetValueResponse, _storage_backend_GetValueRequest__Output, _storage_backend_GetValueResponse__Output>
  OnCreate: MethodDefinition<_storage_backend_OnCreateRequest, _storage_backend_OnCreateResponse, _storage_backend_OnCreateRequest__Output, _storage_backend_OnCreateResponse__Output>
  OnDelete: MethodDefinition<_storage_backend_OnDeleteRequest, _storage_backend_OnDeleteResponse, _storage_backend_OnDeleteRequest__Output, _storage_backend_OnDeleteResponse__Output>
  OnLock: MethodDefinition<_storage_backend_OnLockRequest, _storage_backend_OnLockResponse, _storage_backend_OnLockRequest__Output, _storage_backend_OnLockResponse__Output>
  OnUnlock: MethodDefinition<_storage_backend_OnUnlockRequest, _storage_backend_OnUnlockResponse, _storage_backend_OnUnlockRequest__Output, _storage_backend_OnUnlockResponse__Output>
  OnUpdate: MethodDefinition<_storage_backend_OnUpdateRequest, _storage_backend_OnUpdateResponse, _storage_backend_OnUpdateRequest__Output, _storage_backend_OnUpdateResponse__Output>
}
