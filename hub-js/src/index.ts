import { ApolloServer } from "apollo-server-express";
import express from "express";
import neo4j, { Driver } from "neo4j-driver";
import {
  createTerminus,
  HealthCheck,
  HealthCheckError,
} from "@godaddy/terminus";
import * as http from "http";
import { GraphQLSchema } from "graphql";

import { assertSchemaOnDatabase, getSchemaForMode, HubMode } from "./schema";
import { config } from "./config";
import { logger } from "./logger";
import { ensureCoreStorageTypeInstance } from "./local/resolver/mutation/register-built-in-storage";
import DelegatedStorageService from "./local/storage/service";
import UpdateArgsContainer from "./local/storage/update-args-container";

async function main() {
  logger.info("Using Neo4j database", { endpoint: config.neo4j.endpoint });

  const driver = neo4j.driver(
    config.neo4j.endpoint,
    neo4j.auth.basic(config.neo4j.username, config.neo4j.password)
  );

  const schema = getSchemaForMode(config.hubMode);

  // TODO: Create indexes in Public Hub on DB Populator
  assertSchemaOnDatabase(schema, driver);

  const healthCheck = async () => {
    try {
      return {
        db: await driver.verifyConnectivity(),
      };
    } catch (error) {
      throw new HealthCheckError("health check failed", error);
    }
  };

  const server = await setupHttpServer(schema, driver, healthCheck);
  const { bindPort, bindAddress } = config.graphql;

  logger.info("Starting Hub", { mode: config.hubMode });

  if (config.hubMode === HubMode.Local) {
    await ensureCoreStorageTypeInstance({ driver });
    logger.info(
      "Successfully registered TypeInstance for core backend storage"
    );
  }

  server.listen(bindPort, bindAddress, () => {
    logger.info("GraphQL API is listening", {
      endpoint: `http://${bindAddress}:${bindPort}/graphql`,
    });
  });
}

async function setupHttpServer(
  schema: GraphQLSchema,
  driver: Driver,
  healthCheck: HealthCheck
): Promise<http.Server> {
  const app = express();
  app.use(express.json({ limit: config.express.bodySizeLimit }));

  const delegatedStorage = new DelegatedStorageService(driver);
  const apolloServer = new ApolloServer({
    schema,
    context: () => {
      return {
        driver,
        delegatedStorage,
        updateArgs: new UpdateArgsContainer(),
      };
    },
  });
  await apolloServer.start();
  apolloServer.applyMiddleware({ app });

  const server = http.createServer(app);

  createTerminus(server, {
    healthChecks: { "/healthz": healthCheck },
  });

  return server;
}

(async () => {
  await main();
})();
