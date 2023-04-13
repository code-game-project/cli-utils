package cli

import (
	"errors"
	"regexp"

	"github.com/AlecAivazis/survey/v2"
)

type Validator func(input interface{}) error

func MinLength(n int) Validator {
	return Validator(survey.MinLength(n))
}

func MaxLength(n int) Validator {
	return Validator(survey.MaxLength(n))
}

func Regexp(regexp *regexp.Regexp, errorMsg string) Validator {
	return func(input interface{}) error {
		text, ok := input.(string)
		if !ok {
			return errors.New("value must be a string")
		}
		if !regexp.MatchString(text) {
			return errors.New(errorMsg)
		}
		return nil
	}
}
