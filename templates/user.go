package templates

import (
	"errors"
	"os/user"
	"strings"

	"github.com/code-game-project/cli-utils/exec"
	"github.com/code-game-project/cli-utils/feedback"
)

const FeedbackPkg = feedback.Package("templates")

var ErrNameNotFound = errors.New("couldn't determine username")

// GetUsername tries to determine the name of the current user.
// It looks at the following things in order:
//  1. git config user.name
//  2. currently logged in user of the OS
func GetUsername() (string, error) {
	name, err := exec.ExecuteHidden("git", "config", "user.name")
	if err == nil {
		return strings.TrimSpace(name), nil
	}

	user, err := user.Current()
	if err == nil {
		return strings.TrimSpace(user.Username), nil
	}

	return "", ErrNameNotFound
}
