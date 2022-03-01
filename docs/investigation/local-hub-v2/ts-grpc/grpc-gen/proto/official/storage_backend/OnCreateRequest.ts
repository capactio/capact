// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface OnCreateRequest {
  'typeInstanceId'?: (string);
  'value'?: (Buffer | Uint8Array | string);
  'context'?: (Buffer | Uint8Array | string);
}

export interface OnCreateRequest__Output {
  'typeInstanceId': (string);
  'value': (Buffer);
  'context': (Buffer);
}
