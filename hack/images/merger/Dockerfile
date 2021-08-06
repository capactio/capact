FROM mikefarah/yq:4.11.2

WORKDIR /yamls
USER root

RUN apk add --no-cache "bash=>5.1.0"

COPY merger.sh /

ENTRYPOINT ["/merger.sh"]
