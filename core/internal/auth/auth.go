package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	service = "notesync"
	account = "notion-token"

	notionAPIBase  = "https://api.notion.com/v1"
	notionVersion  = "2022-06-28"
	requestTimeout = 15 * time.Second
)

func Store(token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("token is empty")
	}

	name, err := verify(token)
	if err != nil {
		return "", err
	}

	if err := keyring.Set(service, account, token); err != nil {
		return "", fmt.Errorf("save to keyring: %w", err)
	}
	return name, nil
}

func Get() (string, error) {
	token, err := keyring.Get(service, account)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", errors.New("not authenticated (run `notesync auth`)")
		}
		return "", fmt.Errorf("read from keyring: %w", err)
	}
	return token, nil
}

func Delete() error {
	err := keyring.Delete(service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func verify(token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, notionAPIBase+"/users/me", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Notion-Version", notionVersion)

	client := &http.Client{Timeout: requestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("contacting Notion: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("Notion rejected the token (invalid or revoked)")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Notion returned %d: %s", resp.StatusCode, string(body))
	}

	var me struct {
		Name string `json:"name"`
		Bot  struct {
			Owner struct {
				Type string `json:"type"`
			} `json:"owner"`
		} `json:"bot"`
	}
	if err := json.Unmarshal(body, &me); err != nil {
		return "", fmt.Errorf("parsing Notion response: %w", err)
	}
	if me.Name == "" {
		me.Name = "Notion integration"
	}
	return me.Name, nil
}
