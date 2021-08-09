package action

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/docker/docker/pkg/namesgenerator"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/yaml"
)

func isDNSSubdomain(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("Cannot enforce DNS syntax validation on response of type %T", val)
	}

	validation.IsDNS1123Subdomain(str)
	if msgs := validation.IsDNS1123Subdomain(str); len(msgs) != 0 {
		return fmt.Errorf("%s", strings.Join(msgs, ", "))
	}

	return nil
}

func validatorAdapter(validate func(inputParams string) error) survey.Validator {
	return func(val interface{}) error {
		str, ok := val.(string)
		if !ok {
			return fmt.Errorf("Cannot enforce input parameters JSONSchema validation on response of type %T", val)
		}

		if err := validate(str); err != nil {
			// without new line, error is inlined, example output:
			//
			// X Sorry, your reply was invalid: - TypeInstances "testUpdate":
			//    * required but missing TypeInstance of type cap.type.capactio.capact.validation.update:0.1.0
			return fmt.Errorf("\n%s", err.Error())
		}
		return nil
	}
}

func isYAML(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("Cannot enforce YAML syntax validation on response of type %T", val)
	}

	out := map[string]interface{}{}
	return yaml.Unmarshal([]byte(str), &out)
}

func namespaceQuestion() *survey.Question {
	return &survey.Question{
		Name: "namespace",
		Prompt: &survey.Input{
			Message: "Please type Action namespace: ",
			Default: defaultNamespace,
		},
		Validate: survey.ComposeValidators(survey.Required),
	}
}

func actionNameQuestion(defaultName string) *survey.Question {
	return &survey.Question{
		Name: "name",
		Prompt: &survey.Input{
			Message: "Please type Action name: ",
			Default: defaultName,
		},
		Validate: survey.ComposeValidators(survey.Required, isDNSSubdomain),
	}
}

// generateDNSName returns a DNS-1123 subdomain compliant random name
func generateDNSName() string {
	rand.Seed(time.Now().UnixNano())
	return strings.Replace(namesgenerator.GetRandomName(0), "_", "-", 1)
}
