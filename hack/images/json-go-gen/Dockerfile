FROM node:lts-alpine3.12

ENV GEN_DIR /opt/quicktype-generator
WORKDIR ${GEN_DIR}

RUN npm install -g quicktype@15.0.258

ENTRYPOINT ["quicktype"]

CMD ["--help"]
