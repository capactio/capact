# grpc-proto-loader example

This example shows how to use different libraries to generate a gRPC client. Used libraries:

- [`@grpc/proto-loader`](https://www.npmjs.com/package/@grpc/proto-loader)
- [`ts-proto`](https://github.com/stephenh/ts-proto)

## App layout

- [client-official.ts](client-official.ts) - Showcase usage of gRPC client generated via [`@grpc/proto-loader`](https://www.npmjs.com/package/@grpc/proto-loader).
- [client-official-await.ts](client-official-await.ts) - Showcase async/await usage of gRPC client generated via [`@grpc/proto-loader`](https://www.npmjs.com/package/@grpc/proto-loader). Promises are implemented manually.
- [client-ts-plugin.ts](client-ts-plugin.ts) - Showcase async/await usage of gRPC client generated via [`ts-proto`](https://github.com/stephenh/ts-proto). Promises are provided via [`nice-grpc`](https://www.npmjs.com/package/nice-grpc).

## Generating the clients

Install dependencies:

```bash
npm install
```

### [`@grpc/proto-loader`](https://www.npmjs.com/package/@grpc/proto-loader)

To generate the TypeScript files into [`./proto/official`](./proto/ts-plugin), run:

```bash
$(npm bin)/proto-loader-gen-types --longs=String --enums=String --defaults --oneofs --outDir=proto/official ../../../../../hub-js/proto/storage_backend.proto
```

This is aliased as a npm script:

```bash
npm run build:official-proto
```

### [`ts-proto`](https://github.com/stephenh/ts-proto)

To generate the TypeScript files into [`./proto/ts-plugin`](./proto/ts-plugin), run:

> **NOTE:** The `./proto/ts-plugin` directory needs to exist.

```bash
$(npm bin)/grpc_tools_node_protoc \
  --plugin=protoc-gen-ts_proto=$(npm bin)/protoc-gen-ts_proto \
  --ts_proto_out=./proto/ts-plugin \
  --ts_proto_opt=esModuleInterop=true,outputServices=generic-definitions,useExactTypes=false \
  --proto_path='../../../../../hub-js/proto/' \
  ./../../../../../hub-js/proto/storage_backend.proto
```

This is aliased as a npm script:

```bash
npm run build:plugin-proto
```

### Running example scenario

This simple project demonstrates the different libraries you can use to perform gRPC calls.

1. Build clients:

   ````bash
   npm run build
   ````

2. Start the [secret storage](./../../../../../cmd/secret-storage-backend/README.md) with `dotenv` provider enabled:

   ````bash
   APP_LOGGER_DEV_MODE=true APP_SUPPORTED_PROVIDERS="dotenv" go run ./cmd/secret-storage-backend/main.go
   ````

3. Now run the client by specifying which example you want to run:

   ```bash
   npm run start:client-official
   ```

   ```bash
   npm run start:client-official-await
   ```

   ```bash
   npm run start:client-ts-plugin
   ```
