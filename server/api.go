package server

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/code-game-project/cli-utils/request"
	"github.com/code-game-project/cli-utils/versions"
)

type GameInfo struct {
	Name          string           `json:"name"`
	CGVersion     versions.Version `json:"cg_version"`
	DisplayName   string           `json:"display_name,omitempty"`
	Description   string           `json:"description,omitempty"`
	Version       string           `json:"version,omitempty"`
	RepositoryURL string           `json:"repository_url,omitempty"`
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
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("fetch CGE file: %w", err)
	}
	return file, nil
}

func CreateGame(gameURL string, public, protected bool, config any) (gameID, joinSecret string, err error) {
	type req struct {
		Public    bool `json:"public"`
		Protected bool `json:"protected"`
		Config    any  `json:"config"`
	}
	type response struct {
		GameID     string `json:"game_id"`
		JoinSecret string `json:"join_secret"`
	}
	resp, err := request.PostJSON[response](request.BaseURL("http", gameURL)+"/api/games", req{
		Public:    public,
		Protected: protected,
		Config:    config,
	})
	if err != nil {
		return "", "", fmt.Errorf("create game: %w", err)
	}
	return resp.GameID, resp.JoinSecret, nil
}

func CreatePlayer(gameURL, gameID, username, joinSecret string) (playerID, playerSecret string, err error) {
	type req struct {
		Username   string `json:"username"`
		JoinSecret string `json:"join_secret,omitempty"`
	}
	type response struct {
		PlayerID     string `json:"player_id"`
		PlayerSecret string `json:"player_secret"`
	}
	resp, err := request.PostJSON[response](request.BaseURL("http", gameURL)+"/api/games/"+gameID+"/players", req{
		Username:   username,
		JoinSecret: joinSecret,
	})
	if err != nil {
		return "", "", fmt.Errorf("create player: %w", err)
	}
	return resp.PlayerID, resp.PlayerSecret, nil
}

type Game struct {
	ID        string `json:"id"`
	Players   int    `json:"players"`
	Protected bool   `json:"protected"`
}

func FetchGames(gameURL string, allowProtected bool) (private int, public []Game, err error) {
	type response struct {
		Private int    `json:"private"`
		Public  []Game `json:"public"`
	}
	var query string
	if !allowProtected {
		query = "?protected=false"
	}
	resp, err := request.FetchJSON[response](request.BaseURL("http", gameURL)+"/api/games"+query, 0)
	if err != nil {
		return 0, nil, fmt.Errorf("fetch games: %w", err)
	}
	return resp.Private, resp.Public, nil
}

func FetchGame(gameURL, gameID string) (Game, error) {
	resp, err := request.FetchJSON[Game](request.BaseURL("http", gameURL)+"/api/games/"+gameID, 0)
	if err != nil {
		return Game{}, fmt.Errorf("fetch game: %w", err)
	}
	return resp, nil
}

func FetchPlayers(gameURL, gameID string) (map[string]string, error) {
	resp, err := request.FetchJSON[map[string]string](request.BaseURL("http", gameURL)+"/api/games/"+gameID+"/players", 0)
	if err != nil {
		return nil, fmt.Errorf("fetch player username: %w", err)
	}
	return resp, nil
}

func FetchUsername(gameURL, gameID, playerID string) (string, error) {
	type response struct {
		Username string `json:"username"`
	}
	resp, err := request.FetchJSON[response](request.BaseURL("http", gameURL)+"/api/games/"+gameID+"/players/"+playerID, 1*time.Minute)
	if err != nil {
		return "", fmt.Errorf("fetch player username: %w", err)
	}
	return resp.Username, nil
}
