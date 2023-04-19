package cgfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/code-game-project/cli-utils/feedback"
	"github.com/code-game-project/cli-utils/request"
	"github.com/code-game-project/cli-utils/versions"
)

const FeedbackPkg = feedback.Package("cgfile")

type CodeGameFileData struct {
	GameName    string           `json:"game_name"`
	GameVersion string           `json:"game_version,omitempty"`
	ProjectType string           `json:"project_type"`
	Language    string           `json:"language"`
	LangConfig  map[string]any   `json:"lang_config,omitempty"`
	GameURL     string           `json:"game_url,omitempty"`
	ModVersion  versions.Version `json:"mod_version"`
}

func Load(projectRoot string) (*CodeGameFileData, error) {
	file, err := os.Open(filepath.Join(projectRoot, ".codegame.json"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := &CodeGameFileData{
		LangConfig: make(map[string]any),
	}
	err = json.NewDecoder(file).Decode(data)
	if err != nil {
		return nil, fmt.Errorf("decode .codegame.json: %w", err)
	}
	data.GameURL = request.TrimURL(data.GameURL)

	return data, nil
}

func (c *CodeGameFileData) Write(dir string) error {
	os.MkdirAll(dir, 0o755)

	file, err := os.Create(filepath.Join(dir, ".codegame.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}
