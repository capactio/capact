// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface OnDeleteRequest {
  'typeinstanceId'?: (string);
  'context'?: (Buffer | Uint8Array | string);
}

export interface OnDeleteRequest__Output {
  'typeinstanceId': (string);
  'context': (Buffer);
}
