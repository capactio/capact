// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface OnLockRequest {
  'typeInstanceId'?: (string);
  'context'?: (Buffer | Uint8Array | string);
  'lockedBy'?: (string);
}

export interface OnLockRequest__Output {
  'typeInstanceId': (string);
  'context': (Buffer);
  'lockedBy': (string);
}
