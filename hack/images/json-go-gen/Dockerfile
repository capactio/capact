FROM node:lts-alpine3.12

ENV GEN_DIR /opt/quicktype-generator
WORKDIR ${GEN_DIR}

RUN npm install -g quicktype

ENTRYPOINT ["quicktype"]

CMD ["--help"]

-l go -s schema -o /local/docs/internal/json-generators/assets/quicktype/implementation.go  --src /local/ocf-spec/0.0.1/schema/implementation.json -S /local/ocf-spec/0.0.1/schema/
