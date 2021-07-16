if (process.env.APP_NEO4J_PASSWORD === undefined) {
  throw new Error("APP_NEO4J_PASSWORD not defined");
}

const graphqlBindAddress = process.env.APP_GRAPH_QL_ADDR || ":8080";
const [graphQLAddr, graphQLPort] = graphqlBindAddress.split(":", 2);

export const config = {
  neo4j: {
    endpoint: process.env.APP_NEO4J_ENDPOINT || "bolt://localhost:7687",
    username: process.env.APP_NEO4J_USERNAME || "neo4j",
    password: process.env.APP_NEO4J_PASSWORD,
  },
  graphql: {
    bindAddress: graphQLAddr,
    bindPort: Number(graphQLPort),
  },
  hubMode: process.env.APP_HUB_MODE || "public",
  express: {
    bodySizeLimit: process.env.APP_EXPRESS_BODY_SIZE_LIMIT || "32mb",
  },
};
