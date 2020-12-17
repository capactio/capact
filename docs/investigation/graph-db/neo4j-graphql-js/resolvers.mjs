import fs from "fs";

export const resolvers = {
    InterfaceRevision: {
        // implementationRevisionsCustom: (object, params, context, resolveInfo) => {
        //     // It doesn't work - see issue https://github.com/neo4j-graphql/neo4j-graphql-js/issues/390
        //     return neo4jgraphql(object, params, context, resolveInfo);
        // },

        implementationRevisionsForRequirementsCustom: async (object, params, context, resolveInfo) => {
            fs.writeFileSync('./tmp.js', JSON.stringify(resolveInfo));
            // Transform requirements from array of objects to array of strings
            let typePaths = [];
            if (params && params.filter && params.filter.requirementsSatisfiedBy && params.filter.requirementsSatisfiedBy.length > 0) {
                const reqs = params.filter.requirementsSatisfiedBy;
                typePaths = reqs.map(({typeRef}) => typeRef && typeRef.path)
            }

            const query = `
                MATCH (this:InterfaceRevision)

                // When Implementation doesn't require anything
                CALL{
                WITH this
                MATCH (implRev:ImplementationRevision)-[:IMPLEMENTS]->(this), (implRev)-[:SPECIFIED_BY]->(implRevSpec:ImplementationSpec)
                WHERE NOT (implRevSpec)-[:REQUIRES]->(:ImplementationRequirement)
                RETURN implRev
            
                UNION
            
                // When Implementation has requirements using AnyOf
                WITH this
                MATCH (implRev:ImplementationRevision)-[:IMPLEMENTS]->(this), (implRev)-[:SPECIFIED_BY]->(implRevSpec:ImplementationSpec)-[:REQUIRES]->(:ImplementationRequirement)-[:ONE_OF]->(reqItem:ImplementationRequirementItem)
            
                // TODO: hardcoded typeRefPath - we could use https://stackoverflow.com/questions/51208263/pass-set-of-parameters-to-neo4j-query
                WHERE reqItem.typeRefPath IN $typePaths
                RETURN implRev
                }
            
                RETURN implRev
                `

            return await runQuery(context, query, {
                typePaths
            })
        }
    },


    Query: {
        // If the query and mutation generation is disabled, then this is how you can leverage the neo4j-graphql-js library
        // in your custom resolvers, to automagically return data from database.
        // Note that filter input parameters won't be injected, as with generation disabled the library won't modify the GraphQL schema.
        // interfaceGroups: (object, params, ctx, resolveInfo) => {
        //     return neo4jgraphql(object, params, ctx, resolveInfo);
        // }
    }
}

// copied from neo4j-graphql-js and adjusted a bit
async function runQuery(context, query, cypherParams) {
    let session;

    if (context.neo4jDatabase || context.neo4jBookmarks) {
        const sessionParams = buildSessionParams(context);

        try {
            // connect to the specified database and/or use bookmarks
            // must be using 4.x version of driver
            session = context.driver.session(sessionParams);
        } catch (e) {
            // throw error if bookmark is specified as failure is better than ignoring user provided bookmark
            if (context.neo4jBookmarks) {
                throw new Error(
                    `context.neo4jBookmarks specified, but unable to set bookmark in session object: ${e.message}`
                );
            } else {
                // error - not using a 4.x version of driver!
                // fall back to default database
                session = context.driver.session();
            }
        }
    } else {
        // no database or bookmark specified
        session = context.driver.session();
    }

    let result;

    try {
        result = await session.readTransaction(async tx => {
            const result = await tx.run(query, cypherParams);

            const extractedResult = result.records.map(function (record) {
                return record.get("implRev");
            });

            const gqlResult = extractedResult.map(node => ({
                _id: node.identity.toNumber(),
                ...node.properties
            }))
            // console.log(gqlResult);
            return gqlResult
        });
    } finally {
        session.close();
    }
    return result;
}
