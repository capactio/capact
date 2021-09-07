FROM neo4j:4.2.8

ARG DESTINATION=/apoc.jar
ARG PLUGIN_JAR_URL=https://github.com/neo4j-contrib/neo4j-apoc-procedures/releases/download/4.2.0.5/apoc-4.2.0.5-all.jar

RUN wget -q --timeout 300 --tries 30 --output-document="${DESTINATION}" "${PLUGIN_JAR_URL}"

COPY neo4jlabs-plugins.json /neo4jlabs-plugins.json
COPY docker-entrypoint.sh /docker-entrypoint.sh

ENTRYPOINT ["/sbin/tini", "-g", "--", "/docker-entrypoint.sh"]
CMD ["neo4j"]
