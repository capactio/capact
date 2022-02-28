// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface OnLockRequest {
  'typeinstanceId'?: (string);
  'context'?: (Buffer | Uint8Array | string);
  'lockedBy'?: (string);
}

export interface OnLockRequest__Output {
  'typeinstanceId': (string);
  'context': (Buffer);
  'lockedBy': (string);
}
