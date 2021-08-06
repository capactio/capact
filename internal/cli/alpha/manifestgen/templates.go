package manifestgen

const (
	interfaceGroupManifestTemplate = `
ocfVersion: 0.0.1
revision: 0.1.0
kind: InterfaceGroup
metadata:
  prefix: "cap.interface{{ if .Prefix }}.{{ .Prefix }}{{ end }}"
  name: "{{ .Name }}"
  displayName: "{{ .Name }}"
  description: "{{ .Name }} related Interfaces"
  documentationURL: https://example.com
  supportURL: https://example.com
  iconURL: https://example.com/icon.png
  maintainers:
    - email: dev@example.cop
      name: Example Dev
      url: https://example.com
`

	interfaceManifestTemplate = `ocfVersion: 0.0.1
revision: {{ .Revision }}
kind: Interface
metadata:
  prefix: "cap.interface.{{ .Prefix }}"
  name: "{{ .Name }}"
  displayName: "{{ .Name }}"
  description: "{{ .Name }} action for {{ .Prefix }}"
  documentationURL: https://example.com
  supportURL: https://example.com
  iconURL: https://example.com/icon.png
  maintainers:
    - email: dev@example.cop
      name: Example Dev
      url: https://example.com

spec:
  input:
    parameters:
      input-parameters:
        typeRef:
          path: cap.type.{{ .Prefix }}.{{ .Name }}-input
          revision: 0.1.0
    typeInstances: {}

  output:
    typeInstances:
      config:
        typeRef:
          path: cap.type.{{ .Prefix }}.config
          revision: 0.1.0
`

	typeManifestTemplate = `ocfVersion: 0.0.1
revision: 0.1.0
kind: Type
metadata:
  prefix: "cap.type.{{ .Prefix }}"
  name: {{ .Name }}-input
  displayName: "Input for {{ .Prefix }}.{{ .Name }}"
  description: Input for the "{{ .Prefix }}.{{ .Name }} Action"
  documentationURL: https://example.com
  supportURL: https://example.com
  maintainers:
    - email: dev@example.com
      name: Example Dev
      url: https://example.com
spec:
  jsonSchema:
    value: |-
      {
        "$schema": "http://json-schema.org/draft-07/schema",
        "type": "object",
        "required": [],
        "properties": {
          {{ $length := len .Variables -}}
          {{ range $index, $variable := .Variables -}}
          "{{ $variable.Name }}": {
            "$id": "#/properties/{{ $variable.Name }}",
            "type": "{{ $variable.Type }}",
            "description": "{{ $variable.Description }}"{{if $variable.Default }},
            "default": "{{ $variable.Default }}"{{ end }}
          }{{ if ne $index (add $length -1) }},{{ end }}
          {{ end }}
        }
      }
`

	outputTypeManifestTemplate = `ocfVersion: 0.0.1
revision: 0.1.0
kind: Type
metadata:
  prefix: "cap.type.{{ .Prefix }}"
  name: config
  displayName: "{{.Prefix }} config"
  description: "Type representing a {{ .Prefix }} config"
  documentationURL: https://example.com
  supportURL: https://example.com
  maintainers:
    - email: dev@example.com
      name: Example Dev
      url: https://example.com
spec:
  jsonSchema:
    # Put the properties of your Interface output Type in form of a JSON Schema here:
    value: |-
      {
        "$schema": "http://json-schema.org/draft-07/schema",
        "type": "object",
        "required": [],
        "properties": {
          "example": {
            "$id": "#/properties/example",
            "type": "string",
            "description": "Example field"
          }
        }
      }
`

	terraformImplementationManifestTemplate = `ocfVersion: 0.0.1
revision: {{ .Revision }}
kind: Implementation
metadata:
  prefix: "cap.implementation.{{ .Prefix }}"
  name: {{ .Name }}
  displayName: "{{ .Name }} Action"
  description: "{{ .Name }} Action"
  documentationURL: https://example.com
  supportURL: https://example.com
  maintainers:
    - email: dev@example.com
      name: Example Dev
      url: https://example.com
  license:
    name: "Apache 2.0"

spec:
  appVersion: "1.0.x" # Set the supported application version here
  additionalInput:
    parameters:
      typeRef:
        path: "cap.type.{{ .Prefix }}.{{ .Name }}-input"
        revision: 0.1.0

  outputTypeInstanceRelations:
    config:
      uses:
        - terraform-release

  implements:
    - path: {{if .InterfacePath}}cap.interface.{{ .InterfacePath }}{{else}}"cap.interface..." # Put here the path of the implemented Interface{{end}}
      revision: {{if .InterfaceRevision}}{{ .InterfaceRevision }}{{else}}0.1.0{{end}}

  requires: {{if eq .Provider "aws"}}
    cap.type.aws.auth:
      allOf:
        - name: credentials
          alias: aws-credentials
          revision: 0.1.0{{else if eq .Provider "gcp"}}
    cap.type.gcp.auth:
      allOf:
        - name: service-account
          alias: gcp-sa
          revision: 0.1.0{{else}}{}{{end}}

  imports:
    - interfaceGroupPath: cap.interface.runner.argo
      alias: argo
      methods:
        - name: run
          revision: 0.1.0
    - interfaceGroupPath: cap.interface.templating.jinja2
      alias: jinja2
      methods:
        - name: template
          revision: 0.1.0
    - interfaceGroupPath: cap.interface.runner.terraform
      alias: terraform
      methods:
        - name: apply
          revision: 0.1.0

  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: deploy
        templates:
          - name: deploy
            inputs:
              artifacts:
                - name: input-parameters
                - name: additional-parameters
                  optional: true
            outputs:
              artifacts: []
            steps:
              - - name: fill-default-input
                  capact-action: jinja2.template
                  arguments:
                    artifacts:
                      - name: input-parameters
                        from: "{{"{{"}}inputs.artifacts.input-parameters{{"}}"}}"
                      - name: template
                        raw:
                          # Put the input parameters from the Interface here and set default values for it:
                          data: |
                            my_property: <@ input.my_property | default("default_value") @>
                      - name: configuration
                        raw:
                          data: |
                            prefix: input

              - - name: create-module-args
                  capact-action: jinja2.template
                  arguments:
                    artifacts:
                      - name: input-parameters
                        from: "{{"{{"}}inputs.artifacts.additional-parameters{{"}}"}}"
                      - name: configuration
                        raw:
                          data: |
                            prefix: additionalInput
                      - name: template
                        raw:
                          data: |
                            command: "apply"
                            module:
                              name: "{{ .Name }}"
                              source: "{{ .ModuleSourceURL }}"
                            env: {{if eq .Provider "aws"}}
                              - AWS_ACCESS_KEY_ID=<@ creds.accessKeyID @>
                              - AWS_SECRET_ACCESS_KEY=<@ creds.secretAccessKey @>{{else if eq .Provider "gcp"}}
                              - GOOGLE_PROJECT=<@ creds.project_id @>
                              - GOOGLE_APPLICATION_CREDENTIALS=/additional{{else}}[]{{end}}
                            output:
                              goTemplate: |
                                {{ range $index, $output := .Outputs -}}
                                {{ $output.Name }}: {{"{{"}} .{{ $output.Name }} {{"}}"}}
                                {{ end }}
                            variables: |+
                              {{ range $index, $variable := .Variables -}}
                              <% if additionalInput.{{ $variable.Name }} -%>
                              {{ $variable.Name }} = "<@ additionalInput.{{ $variable.Name }} @>"
                              <%- endif %>

                              {{ end }}
              - - name: fill-parameters
                  capact-action: jinja2.template
                  arguments:
                    artifacts:
                      - name: template
                        from: "{{"{{"}}steps.create-module-args.outputs.artifacts.render{{"}}"}}"
                      - name: input-parameters
                        from: "{{"{{"}}steps.fill-default-input.outputs.artifacts.render{{"}}"}}"
                      - name: configuration
                        raw:
                          data: |
                            prefix: input
              {{ if eq .Provider "gcp" }}
              - - name: convert-gcp-yaml-to-json
                  template: convert-yaml-to-json
                  arguments:
                    artifacts:
                      - name: in
                        from: "{{"{{"}}workflow.outputs.artifacts.gcp-sa{{"}}"}}"
              {{ end }}{{ if .Provider }}
              - - name: fill-creds
                  capact-action: jinja2.template
                  arguments:
                    artifacts:
                      - name: template
                        from: "{{"{{"}}steps.fill-parameters.outputs.artifacts.render{{"}}"}}"
                      - name: input-parameters
                        {{if eq .Provider "aws" -}}
                        from: "{{"{{"}}workflow.outputs.artifacts.aws-credentials{{"}}"}}"
                        {{- else if eq .Provider "gcp" -}}
                        from: "{{"{{"}}workflow.outputs.artifacts.gcp-credentials{{"}}"}}"
                        {{- end}}
                      - name: configuration
                        raw:
                          data: |
                            prefix: creds
              {{ end }}
              - - name: terraform-apply
                  capact-action: terraform.apply
                  capact-outputTypeInstances:
                    - name: terraform-release
                      from: terraform-release
                  arguments:
                    artifacts:
                      - name: input-parameters
                        {{if .Provider -}}
                        from: "{{"{{"}}steps.fill-creds.outputs.artifacts.render{{"}}"}}"
                        {{- else -}}
                        from: "{{"{{"}}steps.fill-parameters.outputs.artifacts.render{{"}}"}}"
                        {{- end}}
                      - name: runner-context
                        from: "{{"{{"}}workflow.outputs.artifacts.runner-context{{"}}"}}"
                        {{- if eq .Provider "gcp"}}
                      - name: additional
                        from: "{{"{{"}}steps.convert-gcp-yaml-to-json.outputs.artifacts.out{{"}}"}}"
                        {{- end}}

              - - name: render-config
                  capact-outputTypeInstances:
                    - name: config
                      from: render
                  capact-action: jinja2.template
                  arguments:
                    artifacts:
                      - name: input-parameters
                        from: "{{"{{"}}steps.terraform-apply.outputs.artifacts.additional{{"}}"}}"
                      - name: configuration
                        raw:
                          data: ""
                      - name: template
                        raw:
                          # You have fill the properties of the output TypeInstance here:
                          data: |
                            property: value
          {{ if eq .Provider "gcp" }}
          - name: convert-yaml-to-json
            inputs:
              artifacts:
                - name: in
                  path: /file
            container:
              image: ghcr.io/capactio/yq:4 # Original image: mikefarah/yq:4
              command: ["sh", "-c"]
              args: ["sleep 1 && yq eval -j -i /file"]
            outputs:
              artifacts:
                - name: out
                  path: /file
          {{ end -}}
`
)
