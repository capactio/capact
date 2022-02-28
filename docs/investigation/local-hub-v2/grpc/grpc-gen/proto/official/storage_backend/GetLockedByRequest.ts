// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface GetLockedByRequest {
  'typeinstanceId'?: (string);
  'context'?: (Buffer | Uint8Array | string);
}

export interface GetLockedByRequest__Output {
  'typeinstanceId': (string);
  'context': (Buffer);
}
