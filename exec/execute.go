package exec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/code-game-project/cli-utils/cli"
	"github.com/code-game-project/cli-utils/feedback"
)

const FeedbackPkg = feedback.Package("exec")

func IsInstalled(programName string) bool {
	_, err := exec.LookPath(programName)
	return err == nil
}

func Execute(programName string, args ...string) error {
	_, err := execute(false, false, programName, args...)
	return err
}

func ExecuteDimmed(programName string, args ...string) error {
	_, err := execute(false, true, programName, args...)
	return err
}

func ExecuteHidden(programName string, args ...string) (string, error) {
	return execute(true, false, programName, args...)
}

func execute(hidden, dimmed bool, programName string, args ...string) (string, error) {
	if !IsInstalled(programName) {
		return "", fmt.Errorf("'%s' ist not installed", programName)
	}
	cmd := exec.Command(programName, args...)

	var out []byte
	var err error

	if !hidden {
		cmd.Stdin = os.Stdin
		if dimmed {
			cli.SetColor(cli.WhiteDim)
			defer cli.ResetColor()

			outP, err := cmd.StdoutPipe()
			if err != nil {
				return "", err
			}
			errP, err := cmd.StderrPipe()
			if err != nil {
				return "", err
			}
			err = cmd.Start()
			if err != nil {
				err = fmt.Errorf("execute '%s %s': %w", programName, strings.Join(args, " "), err)
				return "", err
			}
			go func() {
				io.Copy(cli.Output(), outP)
			}()
			go func() {
				io.Copy(cli.Output(), errP)
			}()
			err = cmd.Wait()
			if err != nil {
				err = fmt.Errorf("execute '%s %s': %w", programName, strings.Join(args, " "), err)
				return "", err
			}
		} else {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
		}
	} else {
		out, err = cmd.CombinedOutput()
	}

	if err != nil {
		err = fmt.Errorf("execute '%s %s': %w", programName, strings.Join(args, " "), err)
	}

	return string(out), err
}
