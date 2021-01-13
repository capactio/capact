declare let process: {
  env: {
    APP_NEO4J_ENDPOINT: string
    APP_NEO4J_USERNAME: string | undefined
    APP_NEO4J_PASSWORD: string

    APP_GRAPHQL_BIND_PORT: number | undefined

    APP_OCH_MODE: string | undefined
  }
};

export default {
  neo4j: {
    endpoint: process.env.APP_NEO4J_ENDPOINT,
    username: process.env.APP_NEO4J_USERNAME || 'neo4j',
    password: process.env.APP_NEO4J_PASSWORD,
  },
  graphql: {
    bindPort: process.env.APP_GRAPHQL_BIND_PORT || 8080,
  },
  ochMode: process.env.APP_OCH_MODE,
};
