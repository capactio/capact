// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface GetValueRequest {
  'typeInstanceId'?: (string);
  'resourceVersion'?: (number);
  'context'?: (Buffer | Uint8Array | string);
}

export interface GetValueRequest__Output {
  'typeInstanceId': (string);
  'resourceVersion': (number);
  'context': (Buffer);
}
