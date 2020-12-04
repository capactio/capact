# Workflow rendering PoC

The target of this was to show a working example on how to render Voltron workflows.

This PoC covers:
- how to reference and call other Interfaces and Implementations
- how to pass arguments and receive outputs from the called Implementations

Not covered in this PoC:
- how Action CR inputs are rendered into the workflow
- how output TypeInstances are uploaded to OCH

## General idea

### Workflow syntax extensions

The following extensions were done to the Implementation definitions. They must be parsed and evaluated by the Voltron engine in the rendering step.

- `.spec.action.args.workflow.entrypoint.templates[].steps[][].action` - defines the interface/implementation to be imported into the workflow
```yaml
action: # optional field
  name: implementation_or_prefix_path # required field
  prefix: some_prefix # optional field
```

### Rendering algorithm

1. Fetch the root Implementation for the Action. Create the Workflow
2. Find all `WorkflowSteps` in the rendered workflow, which have the `Action` field set. If none are found the rendering is complete. If there are some, then foreach:
   - import the implementation based on the `.action.name` property
   - create the workflow for the imported implementation
   - append all templates from the imported workflow to the rendered workflow. Prefix the template names with some unique string
   - remove the `.action` property in the `WorkflowStep`. Set the `.template` property to the entrypoint of the imported workflow
3. Repeat 2.

### Implications

- We are using standard Argo way of passing artifacts, no special syntax is added. Content Creator must know the inputs and output names of the artifacts used in the imported actions. This is not a problem, as they are defined by the required and optional TypeInstances of the imported actions.
- We must rembember prefixing names in the imported actions. Workflows could generate global artifacts and we could have collisions. Open question is how to fetch the TypeInstance artifacts of imported actions (if there is a need for that).
- Open point is how to handle conditional imports. Lets say we have a `jira.install` implementation, which requires a `postgresql.config` TypeInstance. You could provide it to the action or create it, if not provided. We could just use Argo conditions and check, if the `postgresql.config` TypeInstance artifacts is available or not, but maybe we could determine this during rendering instead, to avoid importing unnecesary actions.

## Examples

### Postgres install using Helm runner

In this case we want to create a Postgres install implementation using Helm runner.

`helm.run` implementation does not use any syntax, which has to be rendered. From the Content Creator point of view it exposes an interface, which:
- takes an input artifacts called `helm-args`, specified in the `helm.run` interface
- returns two output artifacts:
  - `helm-release` - specified in the `helm.run` interface
  - `additional` - is a result of templating the template provided in `helm-args`. It's the Content Creator's choice, what will be the content of this artifact

```yaml
# helm.run implementation
  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: helm
        templates:
          - name: helm
            inputs:
              artifacts:
                - name: helm-args
                  path: "/helm-args"
            outputs:
              artifacts:
                - name: helm-release
                  path: "/helm-release"
                - name: additional
                  path: "/additional"
            container:
              image: gcr.io/projectvoltron/helm-runner:de65286
              env:
                - name: RUNNER_INPUT_PATH
                  value: "{{inputs.artifacts.helm-args.path}}"
```

Using this we can create the following `postgres.install` implementation:

```yaml
  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: postgres-install
        templates:
          - name: postgres-install
            outputs:
              # artifact names match with the typeInstance names so they can be referenced, when creating steps for uploading to OCH
              artifacts:
                - name: postgresql
                  from: "{{steps.postgres-install.outputs.artifacts.additional}}"
                - name: helm-release
                  from: "{{steps.postgres-install.outputs.artifacts.helm-release}}"
            steps:
              - - name: create-install-config
                  action:
                    name: cap.implementation.jinja2.template # implementation of a Jinja2 templater (runner?) :D. It templates the 'template' artifact using 'values' and outputs the 'render' artifact . Used for glueing outputs-inputs between steps.
                  arguments:
                    artifacts:
                      - name: values
                        # mock cap.type.database.postgresql.install-input typeInstance artifact, which would normally be provided by Voltron engine in some prefetch step
                        raw:
                          data: |
                            superuser:
                              username: postgres
                              password: s3cr3t
                            defaultDBName: postgres
                      - name: template
                        raw:
                          data: |
                            context:
                              name: "helm-runner-example"
                              dryRun: false
                              timeout: "10m"
                              platform:
                                namespace: "default"
                                serviceAccountName: "helm-runner-example"
                            args:
                              command: "install"
                              generateName: true
                              chart:
                                name: "postgresql"
                                repo: "https://charts.bitnami.com/bitnami"
                              values:
                                image:
                                  pullPolicy: Always
                                postgresqlDatabase: {{ defaultDBName }}
                                postgresqlPassword: {{ superuser.password }}
                              output:{% raw %}
                                directory: "/"
                                helmRelease:
                                  fileName: "helm-release"
                                additional:
                                  fileName: "additional"
                                  value: |-
                                    host: "{{ template "postgresql.fullname" . }}"
                                    port: "{{ template "postgresql.port" . }}"
                                    defaultDBName: "{{ template "postgresql.database" . }}"
                                    superuser:
                                      username: "{{ template "postgresql.username" . }}"
                                      password: "{{ template "postgresql.password" . }}"{% endraw %}

              - - name: postgres-install
                  # helm.run step
                  action:
                    name: cap.implementation.runner.helm.run
                    prefix: helm-install
                  arguments:
                    artifacts:
                      - name: helm-args
                        from: "{{steps.create-install-config.outputs.artifacts.render}}"
```

This is gonna be rendered into the following workflow:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: render-poc
spec:
  entrypoint: postgres-install
  templates:
  - name: postgres-install
    outputs:
      artifacts:
      - from: '{{steps.postgres-install.outputs.artifacts.additional}}'
        name: postgresql
      - from: '{{steps.postgres-install.outputs.artifacts.helm-release}}'
        name: helm-release
    steps:
    - - arguments:
          artifacts:
          - name: values
            raw:
              data: |
                superuser:
                  username: postgres
                  password: s3cr3t
                defaultDBName: postgres
          - name: template
            raw:
              data: |
                context:
                  name: "helm-runner-example"
                  dryRun: false
                  timeout: "10m"
                  platform:
                    namespace: "default"
                    serviceAccountName: "helm-runner-example"
                args:
                  command: "install"
                  generateName: true
                  chart:
                    name: "postgresql"
                    repo: "https://charts.bitnami.com/bitnami"
                  values:
                    image:
                      pullPolicy: Always
                    postgresqlDatabase: {{ defaultDBName }}
                    postgresqlPassword: {{ superuser.password }}
                  output:{% raw %}
                    directory: "/"
                    helmRelease:
                      fileName: "helm-release"
                    additional:
                      fileName: "additional"
                      value: |-
                        host: "{{ template "postgresql.fullname" . }}"
                        port: "{{ template "postgresql.port" . }}"
                        defaultDBName: "{{ template "postgresql.database" . }}"
                        superuser:
                          username: "{{ template "postgresql.username" . }}"
                          password: "{{ template "postgresql.password" . }}"{% endraw %}
        name: create-install-config
        template: c5OVu5-template # .action replaced with .template
    - - arguments:
          artifacts:
          - from: '{{steps.create-install-config.outputs.artifacts.render}}'
            name: helm-args
        name: postgres-install
        template: helm-install-helm  # .action replaced with .template

  # imported from cap.implementation.jinja2.template
  - container:
      args:
      - /template.yml
      - /values.yml
      - --format=yaml
      - -o
      - /render.yml
      image: dinutac/jinja2docker
      name: ""
      resources: {}
    inputs:
      artifacts:
      - name: template
        path: /template.yml
      - name: values
        path: /values.yml
    name: c5OVu5-template
    outputs:
      artifacts:
      - name: render
        path: /render.yml

  # imported from cap.implementation.runner.helm.run
  - container:
      env:
      - name: RUNNER_INPUT_PATH
        value: '{{inputs.artifacts.helm-args.path}}'
      image: gcr.io/projectvoltron/helm-runner:de65286
      name: ""
      resources: {}
    inputs:
      artifacts:
      - name: helm-args
        path: /helm-args
    name: helm-install-helm
    outputs:
      artifacts:
      - name: helm-release
        path: /helm-release
      - name: additional
        path: /additional
```
