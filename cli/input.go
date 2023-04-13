package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

// ErrCanceled indicates that the user has interrupted the programm, e.g. by using Ctrl + C.
var ErrCanceled = errors.New("canceled")

// InputOrExit works like Input but exits with code 2 on ErrCanceled or code 1 on other errors.
func InputOrExit(prompt string, required bool, defaultValue string, validators ...Validator) string {
	i, err := Input(prompt, required, defaultValue, validators...)
	if err != nil {
		if errors.Is(err, ErrCanceled) {
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	return i
}

// Input ask the user to input a line of text.
func Input(prompt string, required bool, defaultValue string, validators ...Validator) (string, error) {
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
		err = ErrCanceled
	}
	return result, err
}

// YesNoOrExit works like YesNo but exits with code 2 on ErrCanceled or code 1 on other errors.
func YesNoOrExit(question string, defaultValue bool) (yes bool) {
	yes, err := YesNo(question, defaultValue)
	if err != nil {
		if errors.Is(err, ErrCanceled) {
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	return yes
}

// YesNo asks the user a yes/no question.
func YesNo(question string, defaultValue bool) (yes bool, err error) {
	err = survey.AskOne(&survey.Confirm{
		Message: question,
		Default: defaultValue,
	}, &yes, survey.WithValidator(survey.Required))
	if err == terminal.InterruptErr {
		err = ErrCanceled
	}
	return yes, err
}

// SelectOrExit works like Select but exits with code 2 on ErrCanceled or code 1 on other errors.
func SelectOrExit(msg string, options []string) int {
	index, err := Select(msg, options)
	if err != nil {
		if errors.Is(err, ErrCanceled) {
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	return index
}

// Select asks the user to select on option. It returns the index of the chosen option.
func Select(msg string, options []string) (int, error) {
	var index int
	err := survey.AskOne(&survey.Select{
		Message: msg,
		Options: options,
	}, &index, survey.WithValidator(survey.Required))
	if err == terminal.InterruptErr {
		err = ErrCanceled
	}
	return index, err
}

// SelectStringOrExit works like SelectString but exits with code 2 on ErrCanceled or code 1 on other errors.
func SelectStringOrExit(msg string, displayOptions []string, options []string) string {
	selected, err := SelectString(msg, displayOptions, options)
	if err != nil {
		if errors.Is(err, ErrCanceled) {
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	return selected
}

// SelectString asks the user to select on option. It returns the entry in options with the chosen index.
// It panics if the length of displayOptions differs from the length of options.
func SelectString(msg string, displayOptions []string, options []string) (string, error) {
	if len(displayOptions) != len(options) {
		panic("Lengths of displayOptions and options don't match")
	}
	var index int
	err := survey.AskOne(&survey.Select{
		Message: msg,
		Options: displayOptions,
	}, &index, survey.WithValidator(survey.Required))
	if err == terminal.InterruptErr {
		err = ErrCanceled
	}
	return options[index], err
}

// MultiSelectOrExit works like SelectString but exits with code 2 on ErrCanceled or code 1 on other errors.
func MultiSelectOrExit(msg string, options []string, selected []int) []bool {
	result, err := MultiSelect(msg, options, selected)
	if err != nil {
		if errors.Is(err, ErrCanceled) {
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	return result
}

// MultiSelect asks the user to select an arbitrary amount of options.
// It returns a boolean slice of length options with the chosen indices set to true.
func MultiSelect(msg string, options []string, selected []int) ([]bool, error) {
	var indices []int
	err := survey.AskOne(&survey.MultiSelect{
		Message: msg,
		Options: options,
		Default: selected,
	}, &indices)
	if err == terminal.InterruptErr {
		err = ErrCanceled
		return nil, err
	}
	chosen := make([]bool, len(options))
	for _, index := range indices {
		chosen[index] = true
	}
	return chosen, nil
}
