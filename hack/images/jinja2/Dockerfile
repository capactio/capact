FROM python:rc-alpine

# Create folders
RUN mkdir /templates/ /variables/

# Set needed env vars
ENV SCRIPTS_DIR /scripts
ENV TEMPLATES_DIR /templates

# Copy extra scripts: embedded render
COPY jinja2-cli/ $SCRIPTS_DIR/jinja2-cli

RUN pip3 install --no-cache-dir pip==21.1.1
RUN pip3 install --no-cache-dir $SCRIPTS_DIR/jinja2-cli[yaml]

ENTRYPOINT ["jinja2"]
