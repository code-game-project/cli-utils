package server

import (
	"fmt"
	"io"
	"time"

	"github.com/code-game-project/cli-utils/request"
)

type GameInfo struct {
	Name          string `json:"name"`
	CGVersion     string `json:"cg_version"`
	DisplayName   string `json:"display_name,omitempty"`
	Description   string `json:"description,omitempty"`
	Version       string `json:"version,omitempty"`
	RepositoryURL string `json:"repository_url,omitempty"`
}

func FetchGameInfo(gameURL string) (GameInfo, error) {
	baseURL := request.BaseURL("http", gameURL)
	info, err := request.FetchJSON[GameInfo](fmt.Sprintf("%s/api/info", baseURL), 10*time.Minute)
	if err != nil {
		return GameInfo{}, fmt.Errorf("fetch game info: %w", err)
	}
	return info, nil
}

func FetchCGEFile(gameURL string) (io.ReadCloser, error) {
	baseURL := request.BaseURL("http", gameURL)
	file, err := request.FetchFile(fmt.Sprintf("%s/api/events", baseURL), 10*time.Minute, false)
	if err != nil {
		return nil, fmt.Errorf("fetch CGE file: %w", err)
	}
	return file, nil
}
