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
async function main() {
    const svc = new DelegatedStorageClient(config_1.PROTO_PATH, config_1.TARGET);
    const provider = 'dotenv'; // or 'aws_secretsmanager';
    const request = {
        typeinstanceId: '123',
        resourceVersion: 1,
        context: Buffer.from(`{"provider":"${provider}"}`)
    };
    const res = await svc.getValue(request).catch((err) => console.error(err));
    if (res) {
        console.log(`(client) Got server response: ${res.value}`);
    }
}
class DelegatedStorageClient {
    constructor(protoPath, target) {
        const packageDefinition = protoLoader.loadSync(protoPath);
        const proto = grpc.loadPackageDefinition(packageDefinition);
        this.client = new proto.storage_backend.StorageBackend(target, grpc.credentials.createInsecure());
    }
    async getValue(req) {
        return new Promise((resolve, reject) => this.client.GetValue(req, (error, res) => {
            if (error)
                return reject(error);
            return resolve(res);
        }));
    }
}
exports.default = DelegatedStorageClient;
const promisify = (fn) => (req) => new Promise((resolve, reject) => {
    fn(req, (err, res) => {
        if (err)
            return reject(err);
        return resolve(res);
    });
});
(async () => {
    await main();
})();
