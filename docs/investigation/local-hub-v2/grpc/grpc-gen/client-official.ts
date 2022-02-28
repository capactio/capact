import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';

import { ProtoGrpcType } from './proto/official/storage_backend';
import { GetValueRequest } from './proto/official/storage_backend/GetValueRequest';
import { PROTO_PATH, TARGET } from './config';

function main() {
  const packageDefinition = protoLoader.loadSync(
    PROTO_PATH
  );
  const proto = grpc.loadPackageDefinition(
    packageDefinition
  ) as unknown as ProtoGrpcType;
  const client = new proto.storage_backend.StorageBackend(
    TARGET,
    grpc.credentials.createInsecure()
  );

  const deadline = new Date();
  deadline.setSeconds(deadline.getSeconds() + 5);
  client.waitForReady(deadline, (error?: Error) => {
    if (error) {
      console.log(`Client connect error: ${error.message}`);
    } else {
      const provider = 'dotenv'; // or 'aws_secretsmanager';
      const request: GetValueRequest = {
        typeinstanceId: '123',
        resourceVersion: 1,
        context: Buffer.from(`{"provider":"${provider}"}`)
      };
      client.GetValue(request, (error, res) => {
        if (error) {
          console.error(error);
          console.error('Is not found: ', error.code == grpc.status.NOT_FOUND);
        } else if (res) {
          console.log(`(client) Got server response: ${res.value}`);
        }
      });
    }
  });
}

main();

