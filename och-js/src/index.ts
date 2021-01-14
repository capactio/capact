import { ApolloServer } from 'apollo-server-express';
import express from 'express';
import neo4j, { Driver } from 'neo4j-driver';
import { createTerminus } from '@godaddy/terminus';
import http from 'http';
import { GraphQLSchema } from 'graphql';

import { getSchemaForMode, assertSchemaOnDatabase } from './schema';
import config from './config';
import logger from './logger';

function setupHttpServer(schema: GraphQLSchema, driver: Driver): http.Server {
  const app = express();

  const apolloServer = new ApolloServer({ schema, context: { driver } });
  apolloServer.applyMiddleware({ app });

  const server = http.createServer(app);

  const healthCheck = () => Promise.resolve();
  createTerminus(server, {
    healthChecks: { '/healthz': healthCheck },
  });

  return server;
}

logger.info(`Using Neo4j database at ${config.neo4j.endpoint}`);

const driver = neo4j.driver(
  config.neo4j.endpoint,
  neo4j.auth.basic(config.neo4j.username, config.neo4j.password),
);

const schema = getSchemaForMode(config.ochMode);

// TODO figure out, who should create the schema on the Neo4j database
assertSchemaOnDatabase(schema, driver);

const server = setupHttpServer(schema, driver);
const port = config.graphql.bindPort;

logger.info(`Starting OCH in ${config.ochMode} mode`);

server.listen(port, () => {
  logger.info(`GraphQL API listening on http://0.0.0.0:${port}...`);
});
