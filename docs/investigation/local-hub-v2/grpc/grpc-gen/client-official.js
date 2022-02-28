"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
Object.defineProperty(exports, "__esModule", { value: true });
const grpc = __importStar(require("@grpc/grpc-js"));
const protoLoader = __importStar(require("@grpc/proto-loader"));
const config_1 = require("./config");
function main() {
    const packageDefinition = protoLoader.loadSync(config_1.PROTO_PATH);
    const proto = grpc.loadPackageDefinition(packageDefinition);
    const client = new proto.storage_backend.StorageBackend(config_1.TARGET, grpc.credentials.createInsecure());
    const deadline = new Date();
    deadline.setSeconds(deadline.getSeconds() + 5);
    client.waitForReady(deadline, (error) => {
        if (error) {
            console.log(`Client connect error: ${error.message}`);
        }
        else {
            const provider = 'dotenv'; // or 'aws_secretsmanager';
            const request = {
                typeinstanceId: '123',
                resourceVersion: 1,
                context: Buffer.from(`{"provider":"${provider}"}`)
            };
            client.GetValue(request, (error, res) => {
                if (error) {
                    console.error(error);
                    console.error('Is not found: ', error.code == grpc.status.NOT_FOUND);
                }
                else if (res) {
                    console.log(`(client) Got server response: ${res.value}`);
                }
            });
        }
    });
}
main();
