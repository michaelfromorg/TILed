package til

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

type memoryCredentialStore struct {
	secrets   map[string]string
	setErr    error
	getErr    error
	deleteErr error
}

func (store *memoryCredentialStore) Set(service, account, secret string) error {
	if store.setErr != nil {
		return store.setErr
	}
	store.secrets[service+"\x00"+account] = secret
	return nil
}

func (store *memoryCredentialStore) Get(service, account string) (string, error) {
	if store.getErr != nil {
		return "", store.getErr
	}
	secret, ok := store.secrets[service+"\x00"+account]
	if !ok {
		return "", keyring.ErrNotFound
	}
	return secret, nil
}

func (store *memoryCredentialStore) Delete(service, account string) error {
	if store.deleteErr != nil {
		return store.deleteErr
	}
	key := service + "\x00" + account
	if _, ok := store.secrets[key]; !ok {
		return keyring.ErrNotFound
	}
	delete(store.secrets, key)
	return nil
}

func TestKeyringConfigRoundTripExcludesSecret(t *testing.T) {
	store := useMemoryCredentialStore(t)
	root := t.TempDir()
	config := Config{
		DataDir:               root,
		SyncToNotion:          true,
		NotionAPIKey:          "must-not-be-written",
		NotionDBID:            "database-id",
		NotionAPIKeyInKeyring: true,
	}

	require.NoError(t, StoreNotionAPIKey(&config))
	require.NotEmpty(t, config.NotionAPIKeyAccount)
	require.NoError(t, SaveConfig(config))

	content, err := os.ReadFile(filepath.Join(root, ".til", "config"))
	require.NoError(t, err)
	assert.NotContains(t, string(content), config.NotionAPIKey)
	assert.Contains(t, string(content), "NOTION_API_KEY_SOURCE=keyring")
	assert.Contains(t, string(content), "NOTION_API_KEY_ACCOUNT="+config.NotionAPIKeyAccount)

	loaded, err := LoadConfig(root)
	require.NoError(t, err)
	assert.Equal(t, config.NotionAPIKey, loaded.NotionAPIKey)
	assert.Equal(t, config.NotionAPIKeyAccount, loaded.NotionAPIKeyAccount)
	assert.True(t, loaded.NotionAPIKeyInKeyring)
	assert.NoError(t, loaded.NotionAPIKeyLoadError)

	require.NoError(t, DeleteNotionAPIKey(loaded))
	assert.Empty(t, store.secrets)
	missing, err := LoadConfig(root)
	require.NoError(t, err)
	assert.Empty(t, missing.NotionAPIKey)
	assert.Error(t, missing.NotionAPIKeyLoadError)
}

func TestStoreNotionAPIKeyReportsKeyringFailure(t *testing.T) {
	store := useMemoryCredentialStore(t)
	store.setErr = errors.New("keychain unavailable")
	config := Config{
		SyncToNotion:          true,
		NotionAPIKey:          "secret",
		NotionAPIKeyInKeyring: true,
	}

	err := StoreNotionAPIKey(&config)
	assert.ErrorContains(t, err, "keychain unavailable")
	assert.Empty(t, config.NotionAPIKeyAccount)
}

func TestSaveKeyringConfigRequiresAccount(t *testing.T) {
	config := Config{
		DataDir:               t.TempDir(),
		SyncToNotion:          true,
		NotionAPIKey:          "secret",
		NotionDBID:            "database-id",
		NotionAPIKeyInKeyring: true,
	}
	assert.ErrorContains(t, SaveConfig(config), "keychain account is required")
}

func useMemoryCredentialStore(t *testing.T) *memoryCredentialStore {
	t.Helper()
	previous := notionCredentials
	store := &memoryCredentialStore{secrets: map[string]string{}}
	notionCredentials = store
	t.Cleanup(func() {
		notionCredentials = previous
	})
	return store
}
