package content

import (
	"fmt"
	"sort"
	"strings"

	"capact.io/capact/pkg/sdk/manifest"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/pkg/errors"
)

// TerraformConfig stores input parameters for Terraform based content generation
type TerraformConfig struct {
	Config

	ModulePath                string
	ModuleSourceURL           string
	InterfacePathWithRevision string
	Provider                  Provider
}

type terraformTemplatingInput struct {
	templatingInput

	InterfacePath     string
	InterfaceRevision string
	ModuleSourceURL   string
	Outputs           []outputVariable
	Provider          Provider
}

// GenerateTerraformManifests generates manifest files for a Terraform module based Implementation
func GenerateTerraformManifests(cfg *TerraformConfig) (map[string]string, error) {
	input, err := getTerraformTemplatingInput(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting templating input")
	}

	cfgs := []*templatingConfig{
		{
			Template: typeManifestTemplate,
			Input:    input,
		},
		{
			Template: terraformImplementationManifestTemplate,
			Input:    input,
		},
	}

	generated, err := generateManifests(cfgs)
	if err != nil {
		return nil, errors.Wrap(err, "while generating manifests")
	}

	result := make(map[string]string, len(generated))

	for _, m := range generated {
		metadata, err := manifest.GetMetadata([]byte(m))
		if err != nil {
			return nil, errors.Wrap(err, "while getting metadata for manifest")
		}

		manifestPath := fmt.Sprintf("%s.%s", metadata.Metadata.Prefix, metadata.Metadata.Name)

		result[manifestPath] = m
	}

	return result, nil
}

func getTerraformTemplatingInput(cfg *TerraformConfig) (*terraformTemplatingInput, error) {
	module, diags := tfconfig.LoadModule(cfg.ModulePath)
	if diags.Err() != nil {
		return nil, errors.Wrap(diags.Err(), "while loading Terraform module")
	}

	var (
		interfacePath     = cfg.InterfacePathWithRevision
		interfaceRevision = "0.1.0"
	)

	pathSlice := strings.Split(cfg.InterfacePathWithRevision, ":")
	if len(pathSlice) == 2 {
		interfacePath = pathSlice[0]
		interfaceRevision = pathSlice[1]
	}

	input := &terraformTemplatingInput{
		templatingInput: templatingInput{
			Name:      cfg.ManifestName,
			Prefix:    cfg.ManifestsPrefix,
			Revision:  cfg.ManifestRevision,
			Variables: make([]inputVariable, 0, len(module.Variables)),
		},
		InterfacePath:     interfacePath,
		InterfaceRevision: interfaceRevision,
		ModuleSourceURL:   cfg.ModuleSourceURL,
		Outputs:           make([]outputVariable, 0, len(module.Outputs)),
		Provider:          cfg.Provider,
	}

	for _, tfVar := range module.Variables {
		// Skip default for now, as there are problems, when it is a multiline string or with doublequotes in it.
		input.Variables = append(input.Variables, inputVariable{
			Name:        tfVar.Name,
			Type:        getTypeFromTerraformType(tfVar.Type),
			Description: tfVar.Description,
		})
	}

	sort.Slice(input.Variables, func(i, j int) bool {
		return input.Variables[i].Name < input.Variables[j].Name
	})

	for _, tfOut := range module.Outputs {
		input.Outputs = append(input.Outputs, outputVariable{
			Name: tfOut.Name,
		})
	}

	sort.Slice(input.Outputs, func(i, j int) bool {
		return input.Outputs[i].Name < input.Outputs[j].Name
	})

	return input, nil
}

// Terraform types: https://www.terraform.io/docs/language/expressions/types.html
func getTypeFromTerraformType(t string) string {
	if strings.HasPrefix(t, "list") || strings.HasPrefix(t, "tuple") {
		return "array"
	}

	switch t {
	case "string":
		return "string"
	case "number":
		return "number"
	case "bool":
		return "boolean"
	case "null":
		return "null"
	}

	return "object"
}

const (
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
