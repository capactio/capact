import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';

import { ProtoGrpcType } from './proto/official/storage_backend';
import { StorageBackendClient } from './proto/official/storage_backend/StorageBackend';
import { GetValueRequest } from './proto/official/storage_backend/GetValueRequest';
import { PROTO_PATH, TARGET } from './config';
import { GetValueResponse__Output } from './proto/official/storage_backend/GetValueResponse';
import { ServiceError } from '@grpc/grpc-js';
import {
  OnCreateRequest,
  OnDeleteRequest,
} from './proto/ts-plugin/storage_backend';
import { OnDeleteResponse__Output } from './proto/official/storage_backend/OnDeleteResponse';
import { OnCreateResponse__Output } from './proto/official/storage_backend/OnCreateResponse';

async function main() {
  const svc = new DelegatedStorageClient(PROTO_PATH, TARGET);

  const provider = 'dotenv'; // or 'aws_secretsmanager';
  const onCreate: OnCreateRequest = {
    typeInstanceId: '1234',
    context: Buffer.from(`{"provider":"${provider}"}`),
    value: Buffer.from(`{"key":"${provider}"}`),
  };
  const { value, ...ti } = onCreate; // extract common id and context to `ti`

  const createRes = await svc
    .onCreate(onCreate)
    .catch((err: ServiceError) => console.error(err));
  if (createRes) {
    console.log(`TypeInstance created: ${createRes.context}`);
  }
  const onGet: GetValueRequest = {
    ...ti,
    resourceVersion: 1,
  };
  const getRes = await svc
    .getValue(onGet)
    .catch((err: ServiceError) => console.error(err));
  if (getRes) {
    console.log(`Fetch TypeInstance: ${getRes.value}`);
  }

  const onDel: OnDeleteRequest = ti;
  await svc
    .onDelete(onDel)
    .then(() => console.info('Deleted TypeInstance'))
    .catch((err: ServiceError) => console.error(err));
}

export default class DelegatedStorageClient {
  private client: StorageBackendClient;

  constructor(protoPath: string, target: string) {
    const packageDefinition = protoLoader.loadSync(protoPath);
    const proto = grpc.loadPackageDefinition(
      packageDefinition
    ) as unknown as ProtoGrpcType;
    this.client = new proto.storage_backend.StorageBackend(
      target,
      grpc.credentials.createInsecure()
    );
  }

  async getValue(
    req: GetValueRequest
  ): Promise<GetValueResponse__Output | undefined> {
    return new Promise((resolve, reject) =>
      this.client.GetValue(req, (error, res) => {
        if (error) return reject(error);
        return resolve(res);
      })
    );
  }

  async onCreate(
    req: OnCreateRequest
  ): Promise<OnCreateResponse__Output | undefined> {
    return new Promise((resolve, reject) =>
      this.client.onCreate(req, (error, res) => {
        if (error) return reject(error);
        return resolve(res);
      })
    );
  }

  async onDelete(
    req: OnDeleteRequest
  ): Promise<OnDeleteResponse__Output | undefined> {
    return new Promise((resolve, reject) =>
      this.client.onDelete(req, (error, res) => {
        if (error) return reject(error);
        return resolve(res);
      })
    );
  }
}

// TODO: Didn't get make the generic approach working.
type Callback<A, B> = (err?: A, res?: B) => void;

const promisify =
  <T, A, B>(
    fn: (req: T, cb: Callback<grpc.ServiceError, B>) => void
  ): ((req: T) => Promise<B | undefined>) =>
  (req: T) =>
    new Promise((resolve, reject) => {
      fn(req, (err?: grpc.ServiceError, res?: B) => {
        if (err) return reject(err);
        return resolve(res);
      });
    });

(async () => {
  await main();
})();
