# Generating Grakn Go client

Because the repository with protobuf definitions is licensed with AGPL
we do not include it here nor the generated code.

To generate it run following commands in the current dir: 

```bash
git clone https://github.com/graknlabs/protocol.git --branch=1.0.7

protoc --go_out=gograkn  --go_opt=Mprotocol/keyspace/Keyspace.proto=capact.io/capact/poc/graph-db/grakn/go-grakn/gograkn -I=protocol protocol/session/*.proto

protoc --go-grpc_out=gograkn  -I=protocol protocol/session/*.proto
```
All warnings can be ignored. 
