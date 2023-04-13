package cli

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

// Input ask the user to input a line of text.
func Input(prompt string, required bool, defaultValue string, validators ...Validator) string {
	opts := make([]survey.AskOpt, 0, len(validators)+1)
	if required {
		opts = append(opts, survey.WithValidator(survey.Required))
	}
	for _, v := range validators {
		opts = append(opts, survey.WithValidator(survey.Validator(v)))
	}

	var result string
	err := survey.AskOne(&survey.Input{
		Message: prompt,
		Default: defaultValue,
	}, &result, opts...)
	if err == terminal.InterruptErr {
		sendInterrupt()
		return ""
	}
	return result
}

// YesNo asks the user a yes/no question.
func YesNo(question string, defaultValue bool) (yes bool) {
	err := survey.AskOne(&survey.Confirm{
		Message: question,
		Default: defaultValue,
	}, &yes, survey.WithValidator(survey.Required))
	if err == terminal.InterruptErr {
		sendInterrupt()
		return false
	}
	return yes
}

// Select asks the user to select on option. It returns the index of the chosen option.
func Select(msg string, options []string) int {
	var index int
	err := survey.AskOne(&survey.Select{
		Message: msg,
		Options: options,
	}, &index, survey.WithValidator(survey.Required))
	if err == terminal.InterruptErr {
		sendInterrupt()
		return 0
	}
	return index
}

// SelectString asks the user to select on option. It returns the entry in options with the chosen index.
// It panics if the length of displayOptions differs from the length of options.
func SelectString(msg string, displayOptions []string, options []string) string {
	if len(displayOptions) != len(options) {
		panic("Lengths of displayOptions and options don't match")
	}
	var index int
	err := survey.AskOne(&survey.Select{
		Message: msg,
		Options: displayOptions,
	}, &index, survey.WithValidator(survey.Required))
	if err == terminal.InterruptErr {
		sendInterrupt()
		return ""
	}
	return options[index]
}

// MultiSelect asks the user to select an arbitrary amount of options.
// It returns a boolean slice of length options with the chosen indices set to true.
func MultiSelect(msg string, options []string, selected []int) []bool {
	var indices []int
	err := survey.AskOne(&survey.MultiSelect{
		Message: msg,
		Options: options,
		Default: selected,
	}, &indices)
	if err == terminal.InterruptErr {
		sendInterrupt()
		return make([]bool, 0)
	}
	chosen := make([]bool, len(options))
	for _, index := range indices {
		chosen[index] = true
	}
	return chosen
}
