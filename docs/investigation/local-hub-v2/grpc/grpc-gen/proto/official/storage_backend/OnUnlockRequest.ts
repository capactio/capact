// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface OnUnlockRequest {
  'typeinstanceId'?: (string);
  'context'?: (Buffer | Uint8Array | string);
}

export interface OnUnlockRequest__Output {
  'typeinstanceId': (string);
  'context': (Buffer);
}
