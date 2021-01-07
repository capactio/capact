declare let process: {
  env: {
    NEO4J_ENDPOINT: string
    NEO4J_USERNAME: string | undefined
    NEO4J_PASSWORD: string

    GRAPHQL_BIND_PORT: number | undefined

    OCH_MODE: string | undefined
  }
};

export default {
  neo4j: {
    endpoint: process.env.NEO4J_ENDPOINT,
    username: process.env.NEO4J_USERNAME || 'neo4j',
    password: process.env.NEO4J_PASSWORD,
  },
  graphql: {
    bindPort: process.env.GRAPHQL_BIND_PORT || 3000,
  },
  ochMode: process.env.OCH_MODE,
};
