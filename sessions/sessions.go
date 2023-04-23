package sessions

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/adrg/xdg"

	"github.com/code-game-project/cli-utils/feedback"
	"github.com/code-game-project/cli-utils/request"
)

const FeedbackPkg = feedback.Package("sessions")

type Session struct {
	GameURL      string `json:"-"`
	PlayerID     string `json:"-"`
	Username     string `json:"username"`
	GameID       string `json:"game_id"`
	PlayerSecret string `json:"player_secret"`
}

var sessionsPath = filepath.Join(xdg.DataHome, "codegame", "sessions")

func init() {
	os.MkdirAll(sessionsPath, 0o755)
}

func NewSession(gameURL, username, gameID, playerID, playerSecret string) Session {
	return Session{
		GameURL:      gameURL,
		Username:     username,
		GameID:       gameID,
		PlayerID:     playerID,
		PlayerSecret: playerSecret,
	}
}

// ListSessions returns a map of all game URLs in the session store mapped to a list of sessions.
func ListSessions() (map[string][]Session, error) {
	gameDirs, err := os.ReadDir(filepath.Join(sessionsPath))
	if err != nil {
		return nil, err
	}

	result := make(map[string][]Session, len(gameDirs))
	for _, dir := range gameDirs {
		if !dir.IsDir() {
			continue
		}
		gameURL := unescape(dir.Name())
		sessions, err := ListSessionsByGame(gameURL)
		if err != nil {
			continue
		}
		result[gameURL] = sessions
	}

	return result, nil
}

// ListSessionsByGame returns a list of sessions for the game.
func ListSessionsByGame(gameURL string) ([]Session, error) {
	sessionFiles, err := os.ReadDir(filepath.Join(sessionsPath, escape(gameURL)))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make([]Session, 0), nil
		}
		return nil, err
	}
	sessions := make([]Session, 0, len(sessionFiles))
	for _, file := range sessionFiles {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		session, err := LoadSession(gameURL, unescape(strings.TrimSuffix(file.Name(), ".json")))
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// ListGames returns a list of all game URLs in the session store.
func ListGames() ([]string, error) {
	gameDirs, err := os.ReadDir(filepath.Join(sessionsPath))
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(gameDirs))
	for _, dir := range gameDirs {
		if !dir.IsDir() {
			continue
		}
		result = append(result, unescape(dir.Name()))
	}

	return result, nil
}

// Load a session from the session store.
func LoadSession(gameURL, playerID string) (Session, error) {
	data, err := os.ReadFile(filepath.Join(sessionsPath, escape(gameURL), escape(playerID)+".json"))
	if err != nil {
		return Session{}, err
	}

	var session Session
	err = json.Unmarshal(data, &session)
	session.GameURL = gameURL
	session.PlayerID = playerID

	return session, err
}

// Save the session to the session store.
func (s Session) Save() error {
	if s.GameURL == "" {
		return errors.New("empty game url")
	}
	if s.PlayerID == "" {
		return errors.New("empty player id")
	}
	dir := filepath.Join(sessionsPath, escape(s.GameURL))
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, escape(s.PlayerID)+".json"), data, 0o644)
}

// Check returns true if the player still exists in the game.
func (s Session) Check() (bool, error) {
	body, status, err := request.Fetch(request.BaseURL("http", s.GameURL)+"/api/games/"+s.GameID+"/players/"+s.PlayerID, "GET", 0, false, nil)
	if err != nil {
		return false, err
	}
	body.Close()
	if status >= 300 {
		if status == 404 {
			return false, nil
		}
		return false, fmt.Errorf("http status: %s", http.StatusText(status))
	}
	return true, nil
}

// Remove the session from the session store.
func (s Session) Remove() error {
	if s.GameURL == "" {
		return nil
	}
	if s.PlayerID == "" {
		return nil
	}
	dir := filepath.Join(sessionsPath, escape(s.GameURL))
	err := os.Remove(filepath.Join(dir, escape(s.PlayerID)+".json"))
	if err != nil {
		return err
	}

	dirs, err := os.ReadDir(dir)
	if err == nil && len(dirs) == 0 {
		os.Remove(dir)
	}
	return nil
}

// Clean removes all sessions that no longer exist on the game server.
func Clean() error {
	allSessions, err := ListSessions()
	if err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}
	var wg sync.WaitGroup
	for _, sessions := range allSessions {
		for _, ses := range sessions {
			s := ses
			wg.Add(1)
			go func() {
				valid, err := s.Check()
				if err != nil {
					feedback.Error(FeedbackPkg, "Failed to delete invalid session: %s", err)
				} else if !valid {
					s.Remove()
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
	return nil
}

func escape(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func unescape(s string) string {
	gameURL, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return s
	}
	return string(gameURL)
}
