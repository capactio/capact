// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface GetValueResponse {
  'value'?: (Buffer | Uint8Array | string);
  '_value'?: "value";
}

export interface GetValueResponse__Output {
  'value'?: (Buffer);
  '_value': "value";
}
