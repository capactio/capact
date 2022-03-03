// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface GetLockedByRequest {
  'typeInstanceId'?: (string);
  'context'?: (Buffer | Uint8Array | string);
}

export interface GetLockedByRequest__Output {
  'typeInstanceId': (string);
  'context': (Buffer);
}
