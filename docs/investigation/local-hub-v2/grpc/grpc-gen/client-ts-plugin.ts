import {
  GetValueRequest,
  OnCreateRequest,
  OnDeleteRequest,
  StorageBackendDefinition
} from './proto/ts-plugin/storage_backend';
import { createChannel, createClient, Client } from 'nice-grpc';
import { TARGET } from './config';
import { ServiceError } from '@grpc/grpc-js';

async function main() {
  const channel = createChannel(TARGET);
  const client: Client<typeof StorageBackendDefinition> = createClient(
    StorageBackendDefinition,
    channel
  );

  const provider = 'dotenv'; // or 'aws_secretsmanager';
  const onCreate: OnCreateRequest = {
    typeinstanceId: '1234',
    context: Buffer.from(`{"provider":"${provider}"}`),
    value: Buffer.from(`{"key":"${provider}"}`)
  };
  const { value, ...ti } = onCreate; // extract common id and context to `ti`

  const createRes = await client.onCreate(onCreate).catch((err: ServiceError) => console.error(err));
  if (createRes) {
    console.log(`TypeInstance created: ${createRes.context}`);
  }

  const onGet: GetValueRequest = {
    ...ti,
    resourceVersion: 1
  };
  const getRes = await client.getValue(onGet).catch((err: ServiceError) => console.error(err));
  if (getRes) {
    console.log(`Fetch TypeInstance: ${getRes.value}`);
  }

  const onDel: OnDeleteRequest = ti;
  await client.onDelete(onDel).then(() => console.info('Deleted TypeInstance')).catch((err: ServiceError) => console.error(err));
}

(async () => {
  await main();
})();
