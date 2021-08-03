package content

import (
	"fmt"

	"capact.io/capact/pkg/sdk/manifest"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/pkg/errors"
)

// TerraformConfig stores input parameters for Terraform based content generation
type TerraformConfig struct {
	Config
	ModulePath    string
	InterfacePath string
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

func getTerraformTemplatingInput(cfg *TerraformConfig) (*templatingInput, error) {
	module, diags := tfconfig.LoadModule(cfg.ModulePath)
	if diags.Err() != nil {
		return nil, errors.Wrap(diags.Err(), "while loading Terraform module")
	}

	input := &templatingInput{
		Name:          cfg.ManifestName,
		Prefix:        cfg.ManifestsPrefix,
		InterfacePath: cfg.InterfacePath,
		Variables:     make([]inputVariable, 0, len(module.Variables)),
		Outputs:       make([]outputVariable, 0, len(module.Outputs)),
	}

	for _, tfVar := range module.Variables {
		// Skip default for now, as there are problems, when it is multiline string or with doublequotes in it.
		input.Variables = append(input.Variables, inputVariable{
			Name:        tfVar.Name,
			Type:        tfVar.Type,
			Description: tfVar.Description,
		})
	}

	for _, tfOut := range module.Outputs {
		input.Outputs = append(input.Outputs, outputVariable{
			Name: tfOut.Name,
		})
	}

	return input, nil
}

const (
	terraformImplementationManifestTemplate = `ocfVersion: 0.0.1
revision: 0.1.0
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
    - path: {{ if .InterfacePath }}cap.interface.{{ .InterfacePath }}{{else}}"cap.interface..." # Put here the path of the implemented Interface{{end}}
      revision: 0.1.0

  requires: {} # You might need to add here a TypeInstance to access an external API:
    #cap.type.aws.auth:
    #  allOf:
    #    - name: credentials
    #      alias: aws-credentials
    #      revision: 0.1.0

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
                              source: "https://example.com/terraform-module.tgz" # Put the URL for the tarball with the Terraform module source code
                            env: []
                            output:
                              goTemplate:
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

              - - name: terraform-apply
                  capact-action: terraform.apply
                  capact-outputTypeInstances:
                    - name: terraform-release
                      from: terraform-release
                  arguments:
                    artifacts:
                      - name: input-parameters
                        from: "{{"{{"}}steps.fill-parameters.outputs.artifacts.render{{"}}"}}"
                      - name: runner-context
                        from: "{{"{{"}}workflow.outputs.artifacts.runner-context{{"}}"}}"

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
`
)
