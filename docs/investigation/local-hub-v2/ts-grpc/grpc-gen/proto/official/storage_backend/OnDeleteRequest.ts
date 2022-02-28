// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface OnDeleteRequest {
  'typeInstanceId'?: (string);
  'context'?: (Buffer | Uint8Array | string);
}

export interface OnDeleteRequest__Output {
  'typeInstanceId': (string);
  'context': (Buffer);
}
