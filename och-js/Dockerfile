FROM node:15 as builder

WORKDIR /app/

COPY . /app/
RUN yarn && \
    yarn build

FROM node:15

COPY --from=builder /app/dist /app/dist
COPY --from=builder /app/node_modules /app/node_modules
COPY graphql /app/graphql
COPY package.json /app/
COPY docker/entrypoint.sh /entrypoint.sh

WORKDIR /app

ENTRYPOINT ["/entrypoint.sh"]

# TODO Node.js 15 has limited support the syntax to import ES module,
# which the Airbnb ESLint preset enforces, e.g. importing directories and files without extension provided.
# We have to add additional flags, to enable experimental support for this in Node.js
CMD ["node", "--experimental-modules", "--es-module-specifier-resolution=node", "dist/index.js"]
