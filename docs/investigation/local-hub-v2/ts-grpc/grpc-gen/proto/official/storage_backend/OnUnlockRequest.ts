// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface OnUnlockRequest {
  'typeInstanceId'?: (string);
  'context'?: (Buffer | Uint8Array | string);
}

export interface OnUnlockRequest__Output {
  'typeInstanceId': (string);
  'context': (Buffer);
}
