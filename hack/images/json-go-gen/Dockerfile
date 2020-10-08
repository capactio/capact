FROM node:lts-alpine3.12

ENV GEN_DIR /opt/quicktype-generator
WORKDIR ${GEN_DIR}

RUN npm install -g quicktype

ENTRYPOINT ["quicktype"]

CMD ["--help"]
