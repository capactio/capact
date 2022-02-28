// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface OnCreateRequest {
  'typeinstanceId'?: (string);
  'value'?: (Buffer | Uint8Array | string);
  'context'?: (Buffer | Uint8Array | string);
}

export interface OnCreateRequest__Output {
  'typeinstanceId': (string);
  'value': (Buffer);
  'context': (Buffer);
}
