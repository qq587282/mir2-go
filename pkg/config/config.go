package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	ServerName string
	ServerIP   string
	ServerPort int

	LoginGate GateConfig
	SelGate   GateConfig
	RunGate   GateConfig
	M2Server  M2Config
	LoginSrv  LoginSrvConfig
	DBServer  DBConfig
}

type GateConfig struct {
	Enable  bool
	IP      string
	Port    int
	MaxConn int
}

type M2Config struct {
	Enable       bool
	IP           string
	Port         int
	DBServerIP   string
	DBServerPort int
	LoginSrvIP   string
	LoginSrvPort int
	GameIP       string
	GamePort     int
}

type LoginSrvConfig struct {
	Enable       bool
	IP           string
	Port         int
	DBServerIP   string
	DBServerPort int
	ServerList   []ServerInfo
}

type ServerInfo struct {
	Index  int
	Name   string
	IP     string
	Port   int
	WebURL string
	Show   bool
	Tag    string
}

type DBConfig struct {
	Enable   bool
	Type     string
	IP       string
	Port     int
	User     string
	Password string
	Database string
}

var GlobalConfig *ServerConfig

func LoadConfig(path string) (*ServerConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return GetDefaultConfig(), nil
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &ServerConfig{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	GlobalConfig = cfg
	return cfg, nil
}

func GetDefaultConfig() *ServerConfig {
	return &ServerConfig{
		ServerName: "Mir2 Server",
		ServerIP:   "0.0.0.0",
		ServerPort: 7000,
		LoginGate: GateConfig{
			Enable:  true,
			IP:      "0.0.0.0",
			Port:    7000,
			MaxConn: 5000,
		},
		SelGate: GateConfig{
			Enable:  true,
			IP:      "0.0.0.0",
			Port:    7100,
			MaxConn: 5000,
		},
		RunGate: GateConfig{
			Enable:  true,
			IP:      "0.0.0.0",
			Port:    7200,
			MaxConn: 10000,
		},
		M2Server: M2Config{
			Enable:       true,
			IP:           "0.0.0.0",
			Port:         6000,
			DBServerIP:   "127.0.0.1",
			DBServerPort: 3306,
		},
		LoginSrv: LoginSrvConfig{
			Enable:       true,
			IP:           "0.0.0.0",
			Port:         5500,
			DBServerIP:   "127.0.0.1",
			DBServerPort: 3306,
			ServerList: []ServerInfo{
				{Index: 0, Name: "Server1", IP: "127.0.0.1", Port: 7200, Show: true, Tag: "NEW"},
			},
		},
	}
}

func SaveConfig(cfg *ServerConfig, path string) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	v.Set("servername", cfg.ServerName)
	v.Set("serverip", cfg.ServerIP)
	v.Set("serverport", cfg.ServerPort)

	if err := v.WriteConfigAs(path); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}
