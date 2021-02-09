# Slow Neo4j queries - notes

### What I did
- Prepared one GraphQL query that consists of multiple ones and run it with hey (https://github.com/rakyll/hey):
  
    ```bash
    hey -c 10 -z 30s -t 0 -m "POST" -H 'Accept-Encoding: gzip, deflate, br' -T "application/json" -A 'application/json' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' -D ./body.json http://localhost:8080/graphql
    ```
  
    One time queries with `curl`:

    ```bash
    curl 'http://localhost:8080/graphql' -H 'Content-Type: application/json' -H 'Accept: application/json' -H 'Connection: keep-alive' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' --data-binary '@body.json' --compressed -w %{time_connect}:%{time_starttransfer}:%{time_total} -o ./output.json
    ```

- Did queries to Public OCH for the following set-ups:

    - Public OCH and Neo4j deployed on kind cluster: used `kubectl port-forward` and executed GraphQL queries directly to Public OCH
    - Public OCH and Neo4j deployed on kind cluster: executed GraphQL queries via Gateway
    - Neo4j deployed on kind cluster, Public OCH ran on local machine
    - Neo4j ran with Docker image on local machine, Public OCH ran on local machine
    
- Observed Grafana dashboard for Neo4j, Public OCH and Gateway Kubernetes Pods before and after changes to resource limits
- Went through our custom Cypher queries to see possible performance issues according to [the article](https://medium.com/neo4j/cypher-query-optimisations-fe0539ce2e5c)
- Set up indexes for most common fields used in Cypher queries (e.g. filters) on the investigation branch:
    As described in https://github.com/neo4j-graphql/neo4j-graphql-js/pull/499, we don't need to use @id directive.

### To do

- Ask on https://community.neo4j.com/
- Investigate very slow first queries to public OCH after its start 


### Summary
- From Grafana dashboard it is clear, that both Neo4j and Public OCH charts have too little CPU and memory limits set. I adjusted them on the investigation branch.
  As we cannot have more than 2 CPUs on CI, we need to introduce separate overrides for local development.
- A few very first queries to public OCH are very slow (more than 20s). After about 3-4 slow queries, queries are much faster (up to 2s). 
  
    ```bash
    ❯ curl 'http://localhost:8080/graphql' -H 'Content-Type: application/json' -H 'Accept: application/json' -H 'Connection: keep-alive' -H 'AuthorizaticzNjcjN0' -H 'NAMESPACE: default' --data-binary '@body2.json' --compressed -w %{time_connect}:%{time_starttransfer}:%{time_total} -o ./output.json
    % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
    Dload  Upload   Total   Spent    Left  Speed
    100 1193k  100 1183k  100 11039  33814    308  0:00:35  0:00:35 --:--:--  333k
    0.002004:0.002557:35.830928 # <--- almost 36 seconds!
    
    ~/repositories/work/go-voltron/docs/investigation/neo4j-slow-queries neo4j-investigation*
    ❯ curl 'http://localhost:8080/graphql' -H 'Content-Type: application/json' -H 'Accept: application/json' -H 'Connection: keep-alive' -H 'AuthorizaticzNjcjN0' -H 'NAMESPACE: default' --data-binary '@body2.json' --compressed -w %{time_connect}:%{time_starttransfer}:%{time_total} -o ./output.json
    % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
    Dload  Upload   Total   Spent    Left  Speed
    100 1193k  100 1183k  100 11039  4971k  46382 --:--:-- --:--:-- --:--:-- 5016k
    0.001427:0.001781:0.238692 # <--- 23 ms
    ```
      
    
- I observed an issue when executing the test query: 
    
    ```json
    {
    "errors": [
        {
          "message": "Resolve function for \"InterfaceRevision.implementationRevisions\" returned undefined",
          "locations": [
            {
              "line": 436,
              "column": 3
            }
          ],
          "path": [
            "interfaceGroups",
            0,
            "interfaces",
            0,
            "revisions",
            0,
            "implementationRevisions"
          ],
          "extensions": {
            "code": "INTERNAL_SERVER_ERROR",
            "exception": {
              "stacktrace": [
                "Error: Resolve function for \"InterfaceRevision.implementationRevisions\" returned undefined",
                "    at /Users/pkosiec/repositories/work/go-voltron/och-js/node_modules/graphql-tools/dist/makeExecutableSchema.js:66:19",
                "    at field.resolve (/Users/pkosiec/repositories/work/go-voltron/och-js/node_modules/graphql-extensions/dist/index.js:134:26)",
                "    at field.resolve (/Users/pkosiec/repositories/work/go-voltron/och-js/node_modules/apollo-server-core/dist/utils/schemaInstrumentation.js:52:26)",
                "    at resolveFieldValueOrError (/Users/pkosiec/repositories/work/go-voltron/och-js/node_modules/graphql/execution/execute.js:467:18)",
                "    at resolveField (/Users/pkosiec/repositories/work/go-voltron/och-js/node_modules/graphql/execution/execute.js:434:16)",
                "    at executeFields (/Users/pkosiec/repositories/work/go-voltron/och-js/node_modules/graphql/execution/execute.js:275:18)",
                "    at collectAndExecuteSubfields (/Users/pkosiec/repositories/work/go-voltron/och-js/node_modules/graphql/execution/execute.js:713:10)",
                "    at completeObjectValue (/Users/pkosiec/repositories/work/go-voltron/och-js/node_modules/graphql/execution/execute.js:703:10)",
                "    at completeValue (/Users/pkosiec/repositories/work/go-voltron/och-js/node_modules/graphql/execution/execute.js:591:12)",
                "    at completeValue (/Users/pkosiec/repositories/work/go-voltron/och-js/node_modules/graphql/execution/execute.js:557:21)"]
            }
          }
      }]
    }   
    ```
    While calling Public OCH via Gateway with the test query, I get this issue 100% of the time. It harder to reproduce the issue with direct call to Public OCH. No error on OCH or in DB logs. This should be investigated further.
    Interestingly, I didn't see the same issue for other resolvers than for `InterfaceRevision.implementationRevisions`. The `InterfaceRevision.implementationRevisions` resolver is a basic generated resolver by `graphql-neo4j-js` with `@relation` directive.


  










---





http://localhost:3000/


- Running Neo4j without limits on dev cluster
- Experimenting with calls without Gateway
  ```bash
  kubectl port-forward -n voltron-system svc/voltron-och-public 3001:80
  ```
  
  Faster
- Observing Grafana dashboard
    Memory: 1.4GB
    CPU: 0.4m
  


## Benchmark

1. Run DB with Public OCH
1. Populate DB
1. Without doing any manual queries, run benchmark

    ```bash
    hey -c 2 -n 10 -t 0 -m "POST" -H 'Accept-Encoding: gzip, deflate, br' -T "application/json" -A 'application/json' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' -D ./body.json http://localhost:8080/graphql
    ```
    
    To make sure the response is correct, see:
    ```bash
    curl 'http://localhost:8080/graphql' -H 'Content-Type: application/json' -H 'Accept: application/json' -H 'Connection: keep-alive' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' --data-binary '@body.json' --compressed -w %{time_connect}:%{time_starttransfer}:%{time_total} -o ./output.json
    ```

### Local Public OCH with local Neo4j on Docker

```bash
hey -c 2 -n 20 -t 0 -m "POST" -H 'Accept-Encoding: gzip, deflate, br' -T "application/json" -A 'application/json' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' -D ./body.json http://localhost:8080/graphql

Summary:
  Total:        28.1253 secs
  Slowest:      23.4186 secs
  Fastest:      0.2733 secs
  Average:      2.7809 secs
  Requests/sec: 0.7111
  
  Total data:   23206077 bytes
  Size/request: 1160303 bytes

Response time histogram:
  0.273 [1]     |■■■
  2.588 [16]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  4.902 [0]     |
  7.217 [1]     |■■■
  9.531 [0]     |
  11.846 [0]    |
  14.160 [0]    |
  16.475 [0]    |
  18.790 [0]    |
  21.104 [1]    |■■■
  23.419 [1]    |■■■


Latency distribution:
  10% in 0.3189 secs
  25% in 0.4095 secs
  50% in 0.4812 secs
  75% in 0.5223 secs
  90% in 18.8233 secs
  95% in 23.4186 secs
  0% in 0.0000 secs

Details (average, fastest, slowest):
  DNS+dialup:   0.0004 secs, 0.2733 secs, 23.4186 secs
  DNS-lookup:   0.0002 secs, 0.0000 secs, 0.0023 secs
  req write:    0.0001 secs, 0.0000 secs, 0.0005 secs
  resp wait:    2.7792 secs, 0.2727 secs, 23.4132 secs
  resp read:    0.0012 secs, 0.0005 secs, 0.0092 secs

Status code distribution:
  [200] 20 responses

```

```bash
❯ hey -c 10 -z 30s -t 0 -m "POST" -H 'Accept-Encoding: gzip, deflate, br' -T "application/json" -A 'application/json' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' -D ./body.json http://localhost:8080/graphql

Summary:
  Total:        31.0410 secs
  Slowest:      3.5632 secs
  Fastest:      0.7850 secs
  Average:      1.6953 secs
  Requests/sec: 5.8632
  
  Total data:   222273142 bytes
  Size/request: 1221281 bytes

Response time histogram:
  0.785 [1]     |■
  1.063 [2]     |■
  1.341 [11]    |■■■■■■
  1.618 [67]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  1.896 [79]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  2.174 [9]     |■■■■■
  2.452 [5]     |■■■
  2.730 [1]     |■
  3.008 [1]     |■
  3.285 [4]     |■■
  3.563 [2]     |■


Latency distribution:
  10% in 1.3785 secs
  25% in 1.5187 secs
  50% in 1.6426 secs
  75% in 1.7874 secs
  90% in 1.9924 secs
  95% in 2.4495 secs
  99% in 3.5632 secs

Details (average, fastest, slowest):
  DNS+dialup:   0.0002 secs, 0.7850 secs, 3.5632 secs
  DNS-lookup:   0.0001 secs, 0.0000 secs, 0.0023 secs
  req write:    0.0001 secs, 0.0000 secs, 0.0004 secs
  resp wait:    1.6759 secs, 0.7841 secs, 3.0227 secs
  resp read:    0.0191 secs, 0.0003 secs, 0.9202 secs

Status code distribution:
  [200] 182 responses
```

## Public OCH on Kind


4.7s - 10s
After bumping values for OCH public - 1,5-2,5s

curl 'http://localhost:3001/graphql' -H 'Content-Type: application/json' -H 'Accept: application/json' -H 'Connection: keep-alive' -H 'AuthorizaticzNjcjN0' -H 'NAMESPACE: default' --data-binary '@body.json' --compressed -w %{time_connect}:%{time_starttransfer}:%{time_total} -o ./output.json
curl 'http://localhost:3001/graphql' -H 'Content-Type: application/json' -H 'Accept: application/json' -H 'Connection: keep-alive' -H 'AuthorizaticzNjcjN0' -H 'NAMESPACE: default' --data-binary '@body2.json' --compressed -w %{time_connect}:%{time_starttransfer}:%{time_total} -o ./output.json


Slow on start - Slow connection from OCH to DB?
First 3 queries ~30s



