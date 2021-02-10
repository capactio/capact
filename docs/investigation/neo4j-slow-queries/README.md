# Slow Neo4j queries - notes

## What I did
- Prepared one GraphQL query that consists of multiple ones and run it with [hey](https://github.com/rakyll/hey):
  
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
    
- Observed Grafana dashboard for Neo4j, Public OCH and Gateway Pods before and after changes to resource limits
- Went through our custom Cypher queries to see possible performance issues according to [the article](https://medium.com/neo4j/cypher-query-optimisations-fe0539ce2e5c)
- Set up indexes for most common fields used in Cypher queries (e.g. filters) on the investigation branch:
    As described [here](https://github.com/neo4j-graphql/neo4j-graphql-js/pull/499), we don't need to use `@id` directive.
- Tried to tune configuration for JavaScript Neo4j Driver connection pool, as well as turn off encryption
    ```javascript
    const driver = neo4j.driver(
        config.neo4j.endpoint,
        neo4j.auth.basic(config.neo4j.username, config.neo4j.password),
        {
            encrypted: false,
            maxConnectionLifetime: 3 * 60 * 60 * 1000, // 3 hours
            maxConnectionPoolSize: 100,
            connectionAcquisitionTimeout: 2 * 60 * 1000, // 120 seconds
            connectionTimeout: 20 * 1000 // 20 seconds
        }
    );  
  ```
- Tried to use HTTP (`neo4j://`) instead of `bolt://` binary protocol, as HTTP is reported by a part of community as faster from `bolt` [#1](https://github.com/neo4j/neo4j-javascript-driver/issues/374) [#2](https://community.neo4j.com/t/barebones-http-requests-much-faster-than-python-neo4j-driver-and-py2neo/3932) [#3](https://github.com/neo4j/neo4j-java-driver/issues/459)  [#4](https://github.com/neo4jrb/activegraph/issues/1381)

## Summary

Unfortunately I ran out of time dedicated for this investigation (8 hours) and I didn't find the root cause. Here are my observations and remarks:

- From Grafana dashboard it is clear, that the requests and limits for Neo4j, Public OCH and Gateway charts can be increased for better performance. I adjusted them on the investigation branch, however, they might be too high.
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
    
    Apparently the neo4j JS driver doesn't initiate connections in the connection pool before any query to DB (see [`_acquire` method for connection pool implementation](https://github.com/neo4j/neo4j-javascript-driver/blob/ab2f6798928e41c4d3c79bc186c351012b81ad5f/src/internal/pool.js#L169) - it is run only on running DB queries. See also the [Driver constructor](https://github.com/neo4j/neo4j-javascript-driver/blob/ab2f6798928e41c4d3c79bc186c351012b81ad5f/src/driver.js#L70), which doesn't initiate any connections). 
  
    I tried to prepare a few sessions to create and keep connections to use for later actual queries, but it didn't help:
    ```javascript
    const driver = neo4j.driver(
       // (...)
    );

    await driver.verifyConnectivity() // it also creates a new connection

    let sessions:Session[] = []
    for (let i=0; i<10;i++) {
        sessions.push(driver.session())
    }
    try {
        const results:Promise<QueryResult>[] = sessions.map(s => {
            return s.run(
                'MATCH (c:ContentMetadata) return c'
            )
        })
        await Promise.all(results)

        for (let s of sessions) {
            await s.close();
        }

        for (let r of results.values()) {
            const res = await r;
            console.log(res.records)
        }
    }
    catch(err) {
        console.log("err", err);
    } finally {
        for (const s of sessions) {
            await s.close()
        }
    }  
    ```
  
    Looks like it's more related to the queries which are executed.

- Even on local machine (Neo4j on Docker image + Public OCH) the query execution times varies. They seem to be quite similar to the execution times on Kind cluster.
  
    20 queries executed with 2 clients:
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
  
    Queries in 30 seconds with 10 clients
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
    While calling Public OCH via Gateway with the test query, I get this issue 100% of the time. It is harder to reproduce the issue with direct call to Public OCH. No error on OCH or in DB logs. This should be investigated further e.g. by checking out the DB queries from `neo4j-graphql-js`.
    Interestingly, I didn't see the same issue for other resolvers than for `InterfaceRevision.implementationRevisions`. The `InterfaceRevision.implementationRevisions` resolver is a basic generated resolver by `graphql-neo4j-js` with `@relation` directive.

## Performance comparison

Before and after changes (resource requests and limits + indexes). I observed a slight increase in performance, however we can't say the issue is solved.

### Process

Machine: Macbook Pro 16 2019 i7, 16GB RAM - Docker: CPU: 5, Memory: 8GB, Swap: 3GB

1. Run new kind cluster from a scratch (`make dev-cluster`)
1. Expose Grafana:
   ```bash
   kubectl port-forward -n monitoring svc/monitoring-grafana 3000:80
   ```
1. Expose OCH public port:
   ```bash
   kubectl port-forward -n voltron-system svc/voltron-och-public 3001:80
   ```
1. Run load generator with 1 concurrent client, for 3 minutes, without any timeout of queries
   ```bash
   hey -c 1 -z 3m -t 0 -m "POST" -T "application/json" -A 'application/json' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' -D ./body-without-implrev.json http://localhost:3001/graphql 
   ```
1. Run load generator with 2 concurrent clients:
   ```bash
   hey -c 2 -z 3m -t 0 -m "POST" -T "application/json" -A 'application/json' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' -D ./body-without-implrev2.json http://localhost:3001/graphql 
   ```   
1. Run the load generator against Gateway
    ```bash
    hey -c 1 -z 3m -t 0 -m "POST" -T "application/json" -A 'application/json' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' -D ./body-without-implrev.json https://gateway.voltron.local/graphql 
    ```

### Before

```bash
hey -c 1 -z 3m -t 0 -m "POST" -T "application/json" -A 'application/json' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' -D ./body-without-implrev.json http://localhost:3001/graphql

Summary:
  Total:        183.2208 secs
  Slowest:      72.7208 secs
  Fastest:      2.8103 secs
  Average:      6.5435 secs
  Requests/sec: 0.1528
  
  Total data:   32560640 bytes
  Size/request: 1162880 bytes

Response time histogram:
  2.810 [1]     |■■
  9.801 [26]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  16.792 [0]    |
  23.783 [0]    |
  30.775 [0]    |
  37.766 [0]    |
  44.757 [0]    |
  51.748 [0]    |
  58.739 [0]    |
  65.730 [0]    |
  72.721 [1]    |■■


Latency distribution:
  10% in 3.5967 secs
  25% in 3.6964 secs
  50% in 3.9934 secs
  75% in 4.1964 secs
  90% in 9.3995 secs
  95% in 72.7208 secs
  0% in 0.0000 secs

Details (average, fastest, slowest):
  DNS+dialup:   0.0001 secs, 2.8103 secs, 72.7208 secs
  DNS-lookup:   0.0001 secs, 0.0000 secs, 0.0028 secs
  req write:    0.0001 secs, 0.0000 secs, 0.0006 secs
  resp wait:    6.5077 secs, 2.7864 secs, 72.6013 secs
  resp read:    0.0356 secs, 0.0170 secs, 0.2002 secs

Status code distribution:
  [200] 28 responses  
```

```bash
❯ hey -c 2 -z 3m -t 0 -m "POST" -T "application/json" -A 'application/json' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' -D ./body-without-implrev2.json http://localhost:3001/graphql

Summary:
  Total:        207.5575 secs
  Slowest:      15.5272 secs
  Fastest:      3.1060 secs
  Average:      4.9153 secs
  Requests/sec: 0.2216
  
  Total data:   38979708 bytes
  Size/request: 1146462 bytes

Response time histogram:
  3.106 [1]     |■■
  4.348 [25]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  5.590 [1]     |■■
  6.832 [1]     |■■
  8.075 [2]     |■■■
  9.317 [1]     |■■
  10.559 [0]    |
  11.801 [2]    |■■■
  13.043 [0]    |
  14.285 [0]    |
  15.527 [1]    |■■


Latency distribution:
  10% in 3.2982 secs
  25% in 3.4984 secs
  50% in 3.8049 secs
  75% in 4.3949 secs
  90% in 11.2413 secs
  95% in 15.5272 secs
  0% in 0.0000 secs

Details (average, fastest, slowest):
  DNS+dialup:   0.0003 secs, 3.1060 secs, 15.5272 secs
  DNS-lookup:   0.0001 secs, 0.0000 secs, 0.0023 secs
  req write:    0.0001 secs, 0.0000 secs, 0.0002 secs
  resp wait:    4.8873 secs, 3.0809 secs, 15.4933 secs
  resp read:    0.0276 secs, 0.0157 secs, 0.1089 secs

Status code distribution:
  [200] 34 responses

Error distribution:
  [5]   Post "http://localhost:3001/graphql": EOF
  [1]   Post "http://localhost:3001/graphql": read tcp [::1]:63705->[::1]:3001: read: connection reset by peer
  [1]   Post "http://localhost:3001/graphql": read tcp [::1]:63706->[::1]:3001: read: connection reset by peer
  [1]   Post "http://localhost:3001/graphql": read tcp [::1]:63734->[::1]:3001: read: connection reset by peer
  [1]   Post "http://localhost:3001/graphql": read tcp [::1]:63750->[::1]:3001: read: connection reset by peer
  [1]   Post "http://localhost:3001/graphql": read tcp [::1]:63771->[::1]:3001: read: connection reset by peer
  [1]   Post "http://localhost:3001/graphql": read tcp [::1]:63790->[::1]:3001: read: connection reset by peer
  [1]   Post "http://localhost:3001/graphql": read tcp [::1]:63808->[::1]:3001: read: connection reset by peer
```

```bash
❯ hey -c 1 -z 3m -t 0 -m "POST" -T "application/json" -A 'application/json' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' -D ./body-without-implrev.json https://gateway.voltron.local/graphql

Summary:
  Total:	184.5540 secs
  Slowest:	16.5373 secs
  Fastest:	7.8061 secs
  Average:	12.3033 secs
  Requests/sec:	0.0813


Response time histogram:
  7.806 [1]	|■■■■■■■■■■■■■
  8.679 [1]	|■■■■■■■■■■■■■
  9.552 [0]	|
  10.425 [1]	|■■■■■■■■■■■■■
  11.299 [3]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  12.172 [1]	|■■■■■■■■■■■■■
  13.045 [2]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■
  13.918 [1]	|■■■■■■■■■■■■■
  14.791 [2]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■
  15.664 [1]	|■■■■■■■■■■■■■
  16.537 [2]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■


Latency distribution:
  10% in 9.9743 secs
  25% in 11.1239 secs
  50% in 12.9130 secs
  75% in 14.9700 secs
  90% in 16.5373 secs
  0% in 0.0000 secs
  0% in 0.0000 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.3346 secs, 7.8061 secs, 16.5373 secs
  DNS-lookup:	0.3338 secs, 0.0000 secs, 5.0063 secs
  req write:	0.0001 secs, 0.0001 secs, 0.0004 secs
  resp wait:	11.5706 secs, 7.5882 secs, 15.2406 secs
  resp read:	0.3979 secs, 0.2178 secs, 0.6265 secs

Status code distribution:
  [200]	15 responses
```

### After

```bash
 hey -c 1 -z 3m -t 0 -m "POST" -T "application/json" -A 'application/json' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' -D ./body-without-implrev.json http://localhost:3001/graphql

Summary:
  Total:	180.9098 secs
  Slowest:	38.0758 secs
  Fastest:	0.8932 secs
  Average:	1.3302 secs
  Requests/sec:	0.7518

  Total data:	158151680 bytes
  Size/request:	1162880 bytes

Response time histogram:
  0.893 [1]	|
  4.611 [134]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  8.330 [0]	|
  12.048 [0]	|
  15.766 [0]	|
  19.484 [0]	|
  23.203 [0]	|
  26.921 [0]	|
  30.639 [0]	|
  34.358 [0]	|
  38.076 [1]	|


Latency distribution:
  10% in 0.9118 secs
  25% in 0.9854 secs
  50% in 1.0055 secs
  75% in 1.0968 secs
  90% in 1.2171 secs
  95% in 1.3982 secs
  99% in 38.0758 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0000 secs, 0.8932 secs, 38.0758 secs
  DNS-lookup:	0.0000 secs, 0.0000 secs, 0.0024 secs
  req write:	0.0001 secs, 0.0000 secs, 0.0034 secs
  resp wait:	1.3075 secs, 0.8734 secs, 38.0409 secs
  resp read:	0.0226 secs, 0.0164 secs, 0.0538 secs

Status code distribution:
  [200]	136 responses
```
```bash
hey -c 2 -z 3m -t 0 -m "POST" -T "application/json" -A 'application/json' -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' -H 'NAMESPACE: default' -D ./body-without-implrev2.json http://localhost:3001/graphql

Summary:
  Total:	190.5885 secs
  Slowest:	4.2461 secs
  Fastest:	1.5986 secs
  Average:	2.5216 secs
  Requests/sec:	0.2361

  Total data:	34393860 bytes
  Size/request:	1146462 bytes

Response time histogram:
  1.599 [1]	|■■■■
  1.863 [0]	|
  2.128 [0]	|
  2.393 [11]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  2.658 [9]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  2.922 [7]	|■■■■■■■■■■■■■■■■■■■■■■■■■
  3.187 [0]	|
  3.452 [1]	|■■■■
  3.717 [0]	|
  3.981 [0]	|
  4.246 [1]	|■■■■


Latency distribution:
  10% in 2.1846 secs
  25% in 2.3108 secs
  50% in 2.4724 secs
  75% in 2.6965 secs
  90% in 2.8839 secs
  95% in 4.2461 secs
  0% in 0.0000 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0002 secs, 1.5986 secs, 4.2461 secs
  DNS-lookup:	0.0001 secs, 0.0000 secs, 0.0020 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0002 secs
  resp wait:	2.4974 secs, 1.5824 secs, 4.2055 secs
  resp read:	0.0239 secs, 0.0161 secs, 0.0412 secs

Status code distribution:
  [200]	30 responses

Error distribution:
  [5]	Post "http://localhost:3001/graphql": EOF
  [1]	Post "http://localhost:3001/graphql": read tcp [::1]:55475->[::1]:3001: read: connection reset by peer
  [1]	Post "http://localhost:3001/graphql": read tcp [::1]:55476->[::1]:3001: read: connection reset by peer
  [1]	Post "http://localhost:3001/graphql": read tcp [::1]:55487->[::1]:3001: read: connection reset by peer
  [1]	Post "http://localhost:3001/graphql": read tcp [::1]:55488->[::1]:3001: read: connection reset by peer
  [1]	Post "http://localhost:3001/graphql": read tcp [::1]:55503->[::1]:3001: read: connection reset by peer
  [1]	Post "http://localhost:3001/graphql": read tcp [::1]:55504->[::1]:3001: read: connection reset by peer
  [1]	Post "http://localhost:3001/graphql": read tcp [::1]:55551->[::1]:3001: read: connection reset by peer
  [1]	Post "http://localhost:3001/graphql": read tcp [::1]:55553->[::1]:3001: read: connection reset by peer
  [1]	Post "http://localhost:3001/graphql": read tcp [::1]:55602->[::1]:3001: read: connection reset by peer
  [1]	Post "http://localhost:3001/graphql": read tcp [::1]:55603->[::1]:3001: read: connection reset by peer
```
```bash
Summary:
  Total:	180.7304 secs
  Slowest:	7.7099 secs
  Fastest:	1.0205 secs
  Average:	1.5316 secs
  Requests/sec:	0.6529


Response time histogram:
  1.020 [1]	|
  1.689 [89]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  2.358 [26]	|■■■■■■■■■■■■
  3.027 [0]	|
  3.696 [0]	|
  4.365 [0]	|
  5.034 [0]	|
  5.703 [0]	|
  6.372 [0]	|
  7.041 [1]	|
  7.710 [1]	|


Latency distribution:
  10% in 1.1295 secs
  25% in 1.2006 secs
  50% in 1.3090 secs
  75% in 1.6861 secs
  90% in 1.9968 secs
  95% in 2.1903 secs
  99% in 7.7099 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0850 secs, 1.0205 secs, 7.7099 secs
  DNS-lookup:	0.0849 secs, 0.0000 secs, 5.0090 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0001 secs
  resp wait:	1.3312 secs, 0.9841 secs, 2.4008 secs
  resp read:	0.1153 secs, 0.0192 secs, 0.2902 secs

Status code distribution:
  [200]	118 responses
```

## Follow-ups

- See actual DB queries which are made from `neo4j-graphql-js` and execute them manually against DB
- Run Public OCH on local machine against Neo4j Desktop (Enterprise DB) for comparison
- Ask on https://community.neo4j.com/
- We can follow this guide: https://neo4j.com/developer/guide-performance-tuning/

    I didn't do that, because I suspected to have an issue with connecting to DB from our components.  

- Create a task to adjust the requests and limits for Public OCH and Neo4j for local development
- Create an issue for the `Resolve function for \"InterfaceRevision.implementationRevisions\" returned undefined` bug
- To improve performance of GraphQL queries from Engine to OCH, communicate with OCH directly.
