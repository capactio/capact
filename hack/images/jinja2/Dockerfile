FROM python:3.9-slim-buster

# Create folders
RUN mkdir /templates/
RUN mkdir /variables/

# Set needed env vars
ENV SCRIPTS_DIR /scripts
ENV TEMPLATES_DIR /templates

# Copy extra scripts: embedded render
COPY jinja2-cli/ $SCRIPTS_DIR/jinja2-cli

RUN pip3 install pip==21.0.1
RUN pip3 install $SCRIPTS_DIR/jinja2-cli[yaml]

ENTRYPOINT ["jinja2"]
