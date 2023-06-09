package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"google.golang.org/protobuf/proto"

	"github.com/code-game-project/cli-utils/versions"
)

type Action string

const (
	ActionInfo      Action = "info"
	ActionCreate    Action = "create"
	ActionUpdate    Action = "update"
	ActionRunClient Action = "run_client"
	ActionRunServer Action = "run_server"
	ActionBuild     Action = "build"
)

type ModuleInfo struct {
	Version         versions.Version
	Actions         []Action                      `json:"actions"`
	LibraryVersions map[string][]versions.Version `json:"library_versions"`
	ProjectTypes    []string                      `json:"project_types"`
}

func execInfo(modulePath string) (ModuleInfo, error) {
	cmd := exec.Command(modulePath, string(ActionInfo))
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return ModuleInfo{}, fmt.Errorf("execute module: %w", err)
	}
	var resp ModuleInfo
	err = json.Unmarshal(output, &resp)
	if err != nil {
		return ModuleInfo{}, fmt.Errorf("decode info response: %w", err)
	}
	if resp.Actions == nil {
		return ModuleInfo{}, fmt.Errorf("invalid info response: missing 'actions' field")
	}
	if resp.LibraryVersions == nil {
		return ModuleInfo{}, fmt.Errorf("invalid info response: missing 'library_versions' field")
	}
	if resp.ProjectTypes == nil {
		return ModuleInfo{}, fmt.Errorf("invalid info response: missing 'application_types' field")
	}
	return resp, nil
}

func (m *Module) ExecInfo(modVersion versions.Version) (ModuleInfo, error) {
	path, err := m.install(modVersion)
	if err != nil {
		return ModuleInfo{}, fmt.Errorf("install module: %w", err)
	}
	return execInfo(path)
}

func (m *Module) ExecCreateClient(gameName, gameURL, language string, cgVersion versions.Version) (modVersion versions.Version, err error) {
	libraryVersion, err := m.findLibraryVersionByCGVersion(ProjectType_CLIENT, cgVersion)
	if err != nil {
		return nil, ErrUnsupportedCodeGameVersion
	}
	modVersion, err = m.findCompatibleModuleVersion(ProjectType_CLIENT, libraryVersion)
	if err != nil {
		return nil, err
	}

	libVersionStr := libraryVersion.String()
	return modVersion, m.execute(modVersion, ProjectType_CLIENT, ActionCreate, &ActionCreateData{
		Language:       language,
		GameName:       gameName,
		ProjectType:    ProjectType_CLIENT,
		GameURL:        &gameURL,
		LibraryVersion: &libVersionStr,
	})
}

func (m *Module) ExecCreateServer(gameName, language string) (modVersion versions.Version, err error) {
	modVersion, err = m.findLatestModuleVersion(ProjectType_SERVER)
	if err != nil {
		return nil, err
	}
	return modVersion, m.execute(modVersion, ProjectType_SERVER, ActionCreate, &ActionCreateData{
		Language:    language,
		GameName:    gameName,
		ProjectType: ProjectType_SERVER,
	})
}

func (m *Module) ExecUpdateClient(language, gameURL string, cgVersion versions.Version) (modVersion versions.Version, err error) {
	libraryVersion, err := m.findLibraryVersionByCGVersion(ProjectType_CLIENT, cgVersion)
	if err != nil {
		return nil, ErrUnsupportedCodeGameVersion
	}
	modVersion, err = m.findCompatibleModuleVersion(ProjectType_CLIENT, libraryVersion)
	if err != nil {
		return nil, err
	}
	libVersionStr := libraryVersion.String()
	return modVersion, m.execute(modVersion, ProjectType_CLIENT, ActionUpdate, &ActionUpdateData{
		ProjectType:    ProjectType_CLIENT,
		Language:       language,
		GameURL:        &gameURL,
		LibraryVersion: &libVersionStr,
	})
}

func (m *Module) ExecUpdateServer(language, gameURL string, cgVersion versions.Version) (modVersion versions.Version, err error) {
	modVersion, err = m.findLatestModuleVersion(ProjectType_CLIENT)
	if err != nil {
		return nil, ErrUnsupportedCodeGameVersion
	}
	return modVersion, m.execute(modVersion, ProjectType_CLIENT, ActionUpdate, &ActionUpdateData{
		ProjectType: ProjectType_CLIENT,
		Language:    language,
	})
}

func (m *Module) ExecRunClient(modVersion versions.Version, gameURL, language, gameID string, playerID, playerSecret *string, spectate bool, args []string) error {
	return m.execute(modVersion, ProjectType_CLIENT, ActionRunClient, &ActionRunClientData{
		GameURL:      gameURL,
		Language:     language,
		Args:         args,
		GameID:       gameID,
		PlayerID:     playerID,
		PlayerSecret: playerSecret,
		Spectate:     spectate,
	})
}

func (m *Module) ExecRunServer(modVersion versions.Version, language string, port *int32, args []string) error {
	return m.execute(modVersion, ProjectType_CLIENT, ActionRunServer, &ActionRunServerData{
		Language: language,
		Args:     args,
		Port:     port,
	})
}

func (m *Module) execute(modVersion versions.Version, projectType ProjectType, action Action, actionData proto.Message) error {
	path, err := m.install(modVersion)
	if err != nil {
		return fmt.Errorf("install module: %w", err)
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
			return fmt.Errorf("encode action data: %w", err)
		}

		file, err := os.CreateTemp(os.TempDir(), "codegame-module-action-data-*")
		if err != nil {
			return fmt.Errorf("create temporary file for action data: %w", err)
		}

		_, err = file.Write(data)
		if err != nil {
			return fmt.Errorf("write action data to temporary file: %w", err)
		}

		cmd.Env = append(cmd.Env, "CG_MODULE_ACTION_DATA_FILE="+file.Name())
	}
	return cmd.Run()
}
