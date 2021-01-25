FROM node:14 as builder

WORKDIR /app/

COPY . /app/
RUN npm install && \
    npm run build

FROM node:14

COPY --from=builder /app/dist /app/dist
COPY --from=builder /app/node_modules /app/node_modules
COPY graphql /app/graphql
COPY package.json /app/
COPY docker/entrypoint.sh /entrypoint.sh

WORKDIR /app

ENTRYPOINT ["/entrypoint.sh"]

CMD ["node", "dist/index.js"]
