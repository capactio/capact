FROM node:lts-alpine3.13

ENV GEN_DIR /opt/quicktype-generator
WORKDIR ${GEN_DIR}

# TODO: Switch to official quicktype app once the issue is resolved: https://github.com/quicktype/quicktype/issues/1590
# Released from https://github.com/pkosiec/quicktype/commit/ec9f3668c11fa36405e9473113f461a14b4e0401
RUN npm install -g @pkosiec/quicktype@15.0.0

ENTRYPOINT ["quicktype"]

CMD ["--help"]
