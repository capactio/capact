FROM node:16-alpine as builder

WORKDIR /app/

COPY package.json package-lock.json /app/
RUN npm install

COPY . /app/
RUN npm run build

FROM node:16-alpine

COPY --from=builder /app/dist /app/dist
COPY --from=builder /app/node_modules /app/node_modules
COPY graphql /app/graphql
COPY package.json package-lock.json /app/
COPY docker/entrypoint.sh /entrypoint.sh

WORKDIR /app

ENTRYPOINT ["/entrypoint.sh"]

CMD ["npm", "start"]
