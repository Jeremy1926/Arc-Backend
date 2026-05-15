package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	clients      = make(map[string]*DB)
	clientsMutex sync.RWMutex
	basePath     string
	registryPath string
)

type ClientMetadata struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	CreatedAt     time.Time              `json:"created_at"`
	Active        bool                   `json:"active"`
	Configuration map[string]interface{} `json:"configuration"`
}

type Registry struct {
	Clients map[string]ClientMetadata `json:"clients"`
	mu      sync.RWMutex
}

var registry *Registry

func InitManager(path string) error {
	basePath = path
	registryPath = filepath.Join(path, "_registry", "clients.json")

	if err := os.MkdirAll(filepath.Dir(registryPath), 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	registry = &Registry{
		Clients: make(map[string]ClientMetadata),
	}

	if err := registry.load(); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load registry: %w", err)
		}
	}

	return nil
}

func (r *Registry) load() error {
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return json.Unmarshal(data, &r.Clients)
}

func (r *Registry) save() error {
	data, err := json.MarshalIndent(r.Clients, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := registryPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, registryPath)
}

func (r *Registry) has(clientID string) bool {
	r.load()
	r.mu.RLock()
	defer r.mu.RUnlock()
	client, exists := r.Clients[clientID]
	if exists {
		return client.Active
	}

	return false
}

func (r *Registry) get(clientID string) (ClientMetadata, bool) {
	r.load()
	r.mu.RLock()
	defer r.mu.RUnlock()
	meta, exists := r.Clients[clientID]
	return meta, exists
}

func (r *Registry) set(clientID string, meta ClientMetadata) error {
	if err := r.load(); err != nil && !os.IsNotExist(err) {
		return err
	}

	r.mu.Lock()
	r.Clients[clientID] = meta
	r.mu.Unlock()

	return r.save()
}

func RegisterClient(clientID, name string) error {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	if _, exists := clients[clientID]; exists {
		return nil
	}

	if registry.has(clientID) {
		return loadClientDatabase(clientID)
	}

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "default"
	}

	clientPath := filepath.Join(basePath, "clients", clientID, serviceName)
	db, err := New(clientPath)
	if err != nil {
		return fmt.Errorf("failed to create client database: %w", err)
	}

	clients[clientID] = db

	metadata := ClientMetadata{
		ID:        clientID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
		Active:    true,
	}

	if err := registry.set(clientID, metadata); err != nil {
		db.Close()
		delete(clients, clientID)
		return fmt.Errorf("failed to register client: %w", err)
	}

	return nil
}

func ConfigureClient(clientID string, config map[string]interface{}) error {
	meta, exists := registry.get(clientID)
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}
	configData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}
	if err := json.Unmarshal(configData, &meta.Configuration); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}
	return registry.set(clientID, meta)
}

func GetClientConfiguration(clientID string) (map[string]interface{}, error) {
	meta, exists := registry.get(clientID)
	if !exists {
		return nil, fmt.Errorf("client not found: %s", clientID)
	}

	return meta.Configuration, nil
}

func GetClientDB(clientID string) (*DB, error) {
	clientsMutex.RLock()
	db, exists := clients[clientID]
	clientsMutex.RUnlock()

	if exists {
		return db, nil
	}

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	if db, exists := clients[clientID]; exists {
		return db, nil
	}

	if !registry.has(clientID) {
		return nil, fmt.Errorf("client not found: %s", clientID)
	}

	if err := loadClientDatabase(clientID); err != nil {
		return nil, err
	}

	return clients[clientID], nil
}

func GetClientMetadata(clientID string) (ClientMetadata, error) {
	meta, exists := registry.get(clientID)
	if !exists {
		return ClientMetadata{}, fmt.Errorf("client not found: %s", clientID)
	}

	return meta, nil
}

func loadClientDatabase(clientID string) error {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "default"
	}

	clientPath := filepath.Join(basePath, "clients", clientID, serviceName)
	db, err := New(clientPath)
	if err != nil {
		return fmt.Errorf("failed to load client database: %w", err)
	}

	clients[clientID] = db
	return nil
}

func ListClients() ([]ClientMetadata, error) {
	if err := registry.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	registry.mu.RLock()
	defer registry.mu.RUnlock()

	var clientList []ClientMetadata
	for _, meta := range registry.Clients {
		clientList = append(clientList, meta)
	}

	return clientList, nil
}

func DeactivateClient(clientID string) error {
	meta, exists := registry.get(clientID)
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	meta.Active = false
	return registry.set(clientID, meta)
}

func ReactivateClient(clientID string) error {
	meta, exists := registry.get(clientID)
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	meta.Active = true
	return registry.set(clientID, meta)
}

func CloseClient(clientID string) error {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	db, exists := clients[clientID]
	if !exists {
		return nil
	}

	if err := db.Close(); err != nil {
		return err
	}

	delete(clients, clientID)
	return nil
}

func CloseAll() error {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	var errors []error

	for clientID, db := range clients {
		if err := db.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close client %s: %w", clientID, err))
		}
	}

	clients = make(map[string]*DB)

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	return nil
}
