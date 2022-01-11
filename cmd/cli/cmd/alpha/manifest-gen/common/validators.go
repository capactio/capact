package common

import (
	"errors"
	"net/mail"
	"net/url"

	"github.com/AlecAivazis/survey/v2"
)

//ManyValidators allow using many validators function in the Survey validator.
func ManyValidators(validateFuns []survey.Validator) func(ans interface{}) error {
	return func(ans interface{}) error {
		for _, fun := range validateFuns {
			if err := fun(ans); err != nil {
				return err
			}
		}
		return nil
	}
}

// ValidateURL validates a URL.
func ValidateURL(ans interface{}) error {
	if str, ok := ans.(string); !ok || (ans != "" && !isURL(str)) {
		return errors.New("URL is not valid")
	}
	return nil
}

// ValidateEmail validates an email.
func ValidateEmail(ans interface{}) error {
	if str, ok := ans.(string); !ok || (ans != "" && !isEmail(str)) {
		return errors.New("email is not valid")
	}
	return nil
}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
