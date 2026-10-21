package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config holds all application configuration
type Config struct {
	App       AppConfig       `yaml:"app"`
	Database  DatabaseConfig  `yaml:"database"`
	HTTP      HTTPConfig      `yaml:"http"`
	TCP       TCPConfig       `yaml:"tcp"`
	UDP       UDPConfig       `yaml:"udp"`
	GRPC      gRPCConfig      `yaml:"grpc"`
	WebSocket WebSocketConfig `yaml:"websocket"`
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string        `yaml:"name"`
	Version     string        `yaml:"version"`
	Environment string        `yaml:"environment"`
	JWTSecret   string        `yaml:"jwt_secret"`
	MaxUsers    int           `yaml:"max_users"`
	Logging     LoggingConfig `yaml:"logging"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Path  string `yaml:"path"`
	Level string `yaml:"level"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type        string `yaml:"type"`
	Path        string `yaml:"path"`
	Timeout     int    `yaml:"timeout"`
	MaxConn     int    `yaml:"max_conn"`
	AutoMigrate bool   `yaml:"auto_migrate"`
}

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	ReadTimeout     int    `yaml:"read_timeout"`
	WriteTimeout    int    `yaml:"write_timeout"`
	ShutdownTimeout int    `yaml:"shutdown_timeout"`
}

// TCPConfig holds TCP server configuration
type TCPConfig struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	MaxConnections    int    `yaml:"max_connections"`
	ReadBufferSize    int    `yaml:"read_buffer_size"`
	WriteBufferSize   int    `yaml:"write_buffer_size"`
	KeepAlive         bool   `yaml:"keep_alive"`
	KeepAliveInterval int    `yaml:"keep_alive_interval"`
}

// UDPConfig holds UDP server configuration
type UDPConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	MaxMessageSize  int    `yaml:"max_message_size"`
	MaxClients      int    `yaml:"max_clients"`
	BroadcastBuffer int    `yaml:"broadcast_buffer"`
}

// gRPCConfig holds gRPC server configuration
type gRPCConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	MaxConns int    `yaml:"max_conns"`
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	ReadBufferSize  int    `yaml:"read_buffer_size"`
	WriteBufferSize int    `yaml:"write_buffer_size"`
	MaxRooms        int    `yaml:"max_rooms"`
	MaxClients      int    `yaml:"max_clients"`
}

// Profile holds server profile configuration
type Profile struct {
	Name   string
	Config *Config
}

// LoadConfig loads configuration from YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to YAML file
func SaveConfig(path string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	err = os.WriteFile(path, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:        "MangaHub",
			Version:     "1.0.0",
			Environment: "development",
			JWTSecret:   "your-secret-key-change-in-production",
			MaxUsers:    100,
			Logging: LoggingConfig{
				Path:  "logs/server.log",
				Level: "info",
			},
		},
		Database: DatabaseConfig{
			Type:        "sqlite3",
			Path:        filepath.Join(os.ExpandEnv("$HOME"), ".mangahub", "data.db"),
			Timeout:     30,
			MaxConn:     10,
			AutoMigrate: true,
		},
		HTTP: HTTPConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			ReadTimeout:     15,
			WriteTimeout:    15,
			ShutdownTimeout: 10,
		},
		TCP: TCPConfig{
			Host:              "0.0.0.0",
			Port:              9090,
			MaxConnections:    50,
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			KeepAlive:         true,
			KeepAliveInterval: 30,
		},
		UDP: UDPConfig{
			Host:            "0.0.0.0",
			Port:            9091,
			MaxMessageSize:  4096,
			MaxClients:      100,
			BroadcastBuffer: 100,
		},
		GRPC: gRPCConfig{
			Host:     "0.0.0.0",
			Port:     9092,
			MaxConns: 100,
		},
		WebSocket: WebSocketConfig{
			Host:            "0.0.0.0",
			Port:            9093,
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			MaxRooms:        50,
			MaxClients:      500,
		},
	}
}

// ProfileManager manages configuration profiles
type ProfileManager struct {
	profiles map[string]*Config
	active   string
}

// NewProfileManager creates a new profile manager
func NewProfileManager() *ProfileManager {
	return &ProfileManager{
		profiles: make(map[string]*Config),
		active:   "default",
	}
}

// RegisterProfile registers a configuration profile
func (pm *ProfileManager) RegisterProfile(name string, config *Config) {
	pm.profiles[name] = config
}

// GetProfile gets a profile by name
func (pm *ProfileManager) GetProfile(name string) (*Config, error) {
	config, ok := pm.profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", name)
	}
	return config, nil
}

// SetActive sets the active profile
func (pm *ProfileManager) SetActive(name string) error {
	_, ok := pm.profiles[name]
	if !ok {
		return fmt.Errorf("profile not found: %s", name)
	}
	pm.active = name
	return nil
}

// GetActive gets the active profile
func (pm *ProfileManager) GetActive() *Config {
	return pm.profiles[pm.active]
}

// GetActiveName gets the active profile name
func (pm *ProfileManager) GetActiveName() string {
	return pm.active
}

// ListProfiles lists all available profiles
func (pm *ProfileManager) ListProfiles() []string {
	var names []string
	for name := range pm.profiles {
		names = append(names, name)
	}
	return names
}
