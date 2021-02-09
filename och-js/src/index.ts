import {ApolloServer} from "apollo-server-express";
import * as express from "express";
import neo4j, {Driver, QueryResult, Result, Session} from "neo4j-driver";
import {
    createTerminus,
    HealthCheck,
    HealthCheckError,
} from "@godaddy/terminus";
import * as http from "http";
import {GraphQLSchema} from "graphql";

import {assertSchemaOnDatabase, getSchemaForMode} from "./schema";
import {config} from "./config";
import {logger} from "./logger";

async function main() {
    logger.info("Using Neo4j database", {endpoint: config.neo4j.endpoint});

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

    await driver.verifyConnectivity()

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


    const schema = getSchemaForMode(config.ochMode);

    assertSchemaOnDatabase(schema, driver);

    const healthCheck = async () => {
        try {
            return {
                db: await driver.verifyConnectivity(),
            };
        } catch (error) {
            throw new HealthCheckError("healthcheck failed", error);
        }
    };

    const server = setupHttpServer(schema, driver, healthCheck);
    const {bindPort, bindAddress} = config.graphql;

    logger.info("Starting OCH", {mode: config.ochMode});

    server.listen(bindPort, bindAddress, () => {
        logger.info("GraphQL API is listening", {
            endpoint: `http://${bindAddress}:${bindPort}`,
        });
    });
}

function setupHttpServer(
    schema: GraphQLSchema,
    driver: Driver,
    healthCheck: HealthCheck
): http.Server {
    const app = express();

    const apolloServer = new ApolloServer({schema, context: {driver}});
    apolloServer.applyMiddleware({app});

    const server = http.createServer(app);

    createTerminus(server, {
        healthChecks: {"/healthz": healthCheck},
    });

    return server;
}

(async () => {
    await main();
})()
