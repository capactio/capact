import { ApolloServer } from "apollo-server-express";
import * as express from "express";
import neo4j, { Driver } from "neo4j-driver";
import {
  createTerminus,
  HealthCheck,
  HealthCheckError,
} from "@godaddy/terminus";
import * as http from "http";
import { GraphQLSchema } from "graphql";

import { assertSchemaOnDatabase, getSchemaForMode, OCHMode } from "./schema";
import { config } from "./config";
import { logger } from "./logger";

function main() {
  logger.info("Using Neo4j database", { endpoint: config.neo4j.endpoint });

  const driver = neo4j.driver(
    config.neo4j.endpoint,
    neo4j.auth.basic(config.neo4j.username, config.neo4j.password)
  );

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
  const { bindPort, bindAddress } = config.graphql;

  logger.info("Starting OCH", { mode: config.ochMode });

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

  const apolloServer = new ApolloServer({ schema, context: { driver } });
  apolloServer.applyMiddleware({ app });

  const server = http.createServer(app);

  createTerminus(server, {
    healthChecks: { "/healthz": healthCheck },
  });

  return server;
}

main();
