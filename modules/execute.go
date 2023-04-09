package modules

import (
	"fmt"
	"os"
	"os/exec"

	"google.golang.org/protobuf/proto"

	"github.com/code-game-project/cli-utils/versions"
)

type Action string

const (
	ActionStatus Action = "status"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionRun    Action = "run"
	ActionBuild  Action = "build"
)

func (m *Module) ExecCreateClient(gameName, gameURL, language string, cgVersion versions.Version) error {
	libraryVersion, err := m.findLibraryVersionByCGVersion(ProjectType_CLIENT, cgVersion)
	if err != nil {
		return ErrUnsupportedCodeGameVersion
	}

	modVersion, err := m.findCompatibleModuleVersion(ProjectType_CLIENT, libraryVersion)
	if err != nil {
		return err
	}
	return m.execute(modVersion, ProjectType_CLIENT, ActionCreate, &ActionCreateData{
		Language:    language,
		GameName:    gameName,
		ProjectType: ProjectType_CLIENT,
		Url:         &gameURL,
	})
}

func (m *Module) ExecCreateServer(gameName, language string) error {
	modVersion, err := m.findLatestModuleVersion(ProjectType_SERVER)
	if err != nil {
		return err
	}
	return m.execute(modVersion, ProjectType_SERVER, ActionCreate, &ActionCreateData{
		Language:    language,
		GameName:    gameName,
		ProjectType: ProjectType_SERVER,
	})
}

func (m *Module) execute(modVersion versions.Version, projectType ProjectType, action Action, actionData proto.Message) error {
	path, err := m.install(modVersion)
	if err != nil {
		return fmt.Errorf("failed to install module: %w", err)
	}

	cmd := exec.Command(path, string(action))
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	var data []byte
	if actionData != nil {
		data, err = proto.Marshal(actionData)
		if err != nil {
			return fmt.Errorf("failed to encode action data: %w", err)
		}

		file, err := os.CreateTemp(os.TempDir(), "codegame-module-action-data-*")
		if err != nil {
			return fmt.Errorf("failed to create temporary file for action data: %w", err)
		}

		_, err = file.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write action data to temporary file: %w", err)
		}

		cmd.Env = append(cmd.Env, "CG_MODULE_ACTION_DATA_FILE="+file.Name())
	}
	return cmd.Run()
}
