package exec

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func IsInstalled(programName string) bool {
	_, err := exec.LookPath(programName)
	return err == nil
}

func Execute(programName string, args ...string) error {
	_, err := execute(false, programName, args...)
	return err
}

func ExecuteHidden(programName string, args ...string) (string, error) {
	return execute(true, programName, args...)
}

func execute(hidden bool, programName string, args ...string) (string, error) {
	if !IsInstalled(programName) {
		return "", fmt.Errorf("'%s' ist not installed", programName)
	}
	cmd := exec.Command(programName, args...)

	var out []byte
	var err error

	if !hidden {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
	} else {
		out, err = cmd.CombinedOutput()
	}

	if err != nil {
		err = fmt.Errorf("failed to execute '%s %s': %w", programName, strings.Join(args, " "), err)
	}

	return string(out), err
}
