"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const storage_backend_1 = require("./proto/ts-plugin/storage_backend");
const nice_grpc_1 = require("nice-grpc");
const config_1 = require("./config");
async function main() {
    const channel = (0, nice_grpc_1.createChannel)(config_1.TARGET);
    const client = (0, nice_grpc_1.createClient)(storage_backend_1.StorageBackendDefinition, channel);
    const provider = 'dotenv'; // or 'aws_secretsmanager';
    const get = {
        typeinstanceId: '1234',
        resourceVersion: 1,
        context: Buffer.from(`{"provider":"${provider}"}`)
    };
    const create = {
        typeinstanceId: get.typeinstanceId,
        context: get.context,
        value: Buffer.from(`{"key":"${provider}"}`)
    };
    const createRes = await client.onCreate(create).catch((err) => console.error(err));
    if (createRes) {
        console.log(`Got server response: ${createRes.context}`);
    }
    await client.getValue(get).then((res) => console.log(`Got server response: ${res.value}`)).catch((err) => console.error(err));
    await client.onDelete({
        typeinstanceId: get.typeinstanceId,
        context: get.context
    }).then(() => console.info("deleted TypeInstance"));
}
(async () => {
    await main();
})();
