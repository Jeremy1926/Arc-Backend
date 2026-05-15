package util

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	managersIndex     map[string]string
	managersIndexMu   sync.RWMutex
	managersIndexPath string
	indexSalt         = []byte{
		0x4a, 0x8f, 0x2c, 0x91, 0x7d, 0x3e, 0x5b, 0x0a,
		0x63, 0xc0, 0x1a, 0x7e, 0x45, 0xb9, 0xd6, 0x83,
	}
)

func hashEmail(email string) string {
	h := sha256.New()
	h.Write(indexSalt)
	h.Write([]byte(email))
	return hex.EncodeToString(h.Sum(nil))
}

func InitManagersIndex(basePath string) error {
	managersIndexPath = filepath.Join(basePath, "_registry", "managers_idx.json")
	managersIndex = make(map[string]string)

	data, err := os.ReadFile(managersIndexPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(managersIndexPath), 0755); err != nil {
				return fmt.Errorf("failed to create registry directory: %w", err)
			}
			data = []byte("{}")
		} else {
			return err
		}
	}

	return json.Unmarshal(data, &managersIndex)
}

func saveManagersIndex() error {
	managersIndexMu.RLock()
	data, err := json.MarshalIndent(managersIndex, "", "  ")
	managersIndexMu.RUnlock()
	if err != nil {
		return err
	}

	tmpPath := managersIndexPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, managersIndexPath)
}

func IndexManager(email, clientID string) error {
	managersIndexMu.Lock()
	managersIndex[hashEmail(email)] = clientID
	managersIndexMu.Unlock()
	return saveManagersIndex()
}

func RemoveManagerIndex(email string) error {
	managersIndexMu.Lock()
	delete(managersIndex, hashEmail(email))
	managersIndexMu.Unlock()
	return saveManagersIndex()
}

func LookupManagerClient(email string) (string, bool) {
	managersIndexMu.RLock()
	defer managersIndexMu.RUnlock()
	clientID, ok := managersIndex[hashEmail(email)]
	return clientID, ok
}
