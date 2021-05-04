FROM node:lts-alpine3.13

WORKDIR /opt/grapqhl-schema-linter

RUN apk add --no-cache bash

RUN npm install -g graphql-schema-linter@2.0.1

COPY lint-multiple-files.sh .

ENTRYPOINT ["/opt/grapqhl-schema-linter/lint-multiple-files.sh"]
