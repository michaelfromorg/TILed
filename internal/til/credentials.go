package til

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

const notionCredentialService = "TILed"

type credentialStore interface {
	Set(service, account, secret string) error
	Get(service, account string) (string, error)
	Delete(service, account string) error
}

type systemCredentialStore struct{}

func (systemCredentialStore) Set(service, account, secret string) error {
	return keyring.Set(service, account, secret)
}

func (systemCredentialStore) Get(service, account string) (string, error) {
	return keyring.Get(service, account)
}

func (systemCredentialStore) Delete(service, account string) error {
	return keyring.Delete(service, account)
}

var notionCredentials credentialStore = systemCredentialStore{}

func StoreNotionAPIKey(config *Config) error {
	if config == nil {
		return errors.New("configuration is required")
	}
	if !config.SyncToNotion {
		return errors.New("Notion synchronization is disabled")
	}
	if !config.NotionAPIKeyInKeyring {
		return errors.New("OS keychain storage is disabled")
	}
	if strings.TrimSpace(config.NotionAPIKey) == "" {
		return errors.New("Notion API key cannot be empty")
	}
	generatedAccount := false
	if config.NotionAPIKeyAccount == "" {
		account, err := newNotionCredentialAccount()
		if err != nil {
			return err
		}
		config.NotionAPIKeyAccount = account
		generatedAccount = true
	}
	if strings.ContainsAny(config.NotionAPIKeyAccount, "\r\n") {
		return errors.New("Notion keychain account cannot contain a newline")
	}
	if err := notionCredentials.Set(
		notionCredentialService,
		config.NotionAPIKeyAccount,
		config.NotionAPIKey,
	); err != nil {
		if generatedAccount {
			config.NotionAPIKeyAccount = ""
		}
		return fmt.Errorf("store Notion API key in OS keychain: %w", err)
	}
	config.NotionAPIKeyLoadError = nil
	return nil
}

func DeleteNotionAPIKey(config Config) error {
	if !config.NotionAPIKeyInKeyring || config.NotionAPIKeyAccount == "" {
		return nil
	}
	if err := notionCredentials.Delete(
		notionCredentialService,
		config.NotionAPIKeyAccount,
	); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("delete Notion API key from OS keychain: %w", err)
	}
	return nil
}

func loadNotionAPIKey(config *Config) {
	if !config.SyncToNotion || !config.NotionAPIKeyInKeyring {
		return
	}
	if config.NotionAPIKeyAccount == "" {
		config.NotionAPIKeyLoadError = errors.New(
			"Notion keychain account is missing; run 'til config edit'",
		)
		return
	}

	secret, err := notionCredentials.Get(
		notionCredentialService,
		config.NotionAPIKeyAccount,
	)
	if err != nil {
		config.NotionAPIKeyLoadError = fmt.Errorf(
			"load Notion API key from OS keychain: %w",
			err,
		)
		return
	}
	if strings.TrimSpace(secret) == "" {
		config.NotionAPIKeyLoadError = errors.New(
			"Notion API key in OS keychain is empty; run 'til config edit'",
		)
		return
	}
	config.NotionAPIKey = secret
	config.NotionAPIKeyLoadError = nil
}

func newNotionCredentialAccount() (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate Notion keychain account: %w", err)
	}
	return "notion-" + hex.EncodeToString(random), nil
}
