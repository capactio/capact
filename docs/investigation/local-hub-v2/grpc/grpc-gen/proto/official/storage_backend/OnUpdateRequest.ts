// Original file: ../../../../../hub-js/proto/storage_backend.proto


export interface OnUpdateRequest {
  'typeinstanceId'?: (string);
  'newResourceVersion'?: (number);
  'newValue'?: (Buffer | Uint8Array | string);
  'context'?: (Buffer | Uint8Array | string);
  '_context'?: "context";
}

export interface OnUpdateRequest__Output {
  'typeinstanceId': (string);
  'newResourceVersion': (number);
  'newValue': (Buffer);
  'context'?: (Buffer);
  '_context': "context";
}
