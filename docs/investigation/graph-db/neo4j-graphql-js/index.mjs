import {ApolloServer} from 'apollo-server';
import neo4j from 'neo4j-driver';
import {makeAugmentedSchema} from 'neo4j-graphql-js';
import {resolvers} from "./resolvers.mjs";
import {readFileSync} from 'fs';

const typeDefs = readFileSync('./graphql/schema.graphql', 'utf-8')

const schema = makeAugmentedSchema({
    typeDefs,
    resolvers,
    config: {
        query: true,
        mutation: true
        // Apart from boolean value, you can exclude specific types from query or mutation generation - for example:
        // query: {
        //     exclude: ["InterfaceRevision"]
        // },
    }
});

const driver = neo4j.driver(
    'bolt://localhost:7687',
    neo4j.auth.basic('neo4j', 'root')
);

const server = new ApolloServer({schema, context: {driver}});

server.listen(3000, '0.0.0.0').then(({url}) => {
    console.log(`GraphQL API ready at ${url}`);
});
