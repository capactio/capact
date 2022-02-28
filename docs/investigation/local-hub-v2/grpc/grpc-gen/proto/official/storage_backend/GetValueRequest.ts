// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface GetValueRequest {
  'typeinstanceId'?: (string);
  'resourceVersion'?: (number);
  'context'?: (Buffer | Uint8Array | string);
}

export interface GetValueRequest__Output {
  'typeinstanceId': (string);
  'resourceVersion': (number);
  'context': (Buffer);
}
