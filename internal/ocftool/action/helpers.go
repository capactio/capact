package action

import (
	"fmt"
	"io"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hokaccha/go-prettyjson"
	"k8s.io/apimachinery/pkg/util/validation"
)

func isDNSSubdomain(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("cannot enforce DNS syntax validation on response of type %T", val)
	}

	validation.IsDNS1123Subdomain(str)
	if msgs := validation.IsDNS1123Subdomain(str); len(msgs) != 0 {
		return fmt.Errorf("%s", strings.Join(msgs, ", "))
	}

	return nil
}

func isYAML(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("cannot enforce YAML syntax validation on response of type %T", val)
	}

	out := map[string]interface{}{}
	return yaml.Unmarshal([]byte(str), &out)
}

func toString(in bool) string {
	if in {
		return "true"
	}
	return "false"
}

func namespaceQuestion() *survey.Question {
	return &survey.Question{
		Name: "namespace",
		Prompt: &survey.Input{
			Message: "Please type Action namespace",
			Default: "default",
		},
		Validate: survey.ComposeValidators(survey.Required),
	}
}

func actionNameQuestion(defaultName string) *survey.Question {
	return &survey.Question{
		Name: "name",
		Prompt: &survey.Input{
			Message: "Please type Action name",
			Default: defaultName,
		},
		Validate: survey.ComposeValidators(survey.Required, isDNSSubdomain),
	}
}

func printJSON(in interface{}, w io.Writer) error {
	out, err := prettyjson.Marshal(in)
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	return err
}

func printYAML(in interface{}, w io.Writer) error {
	out, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}
