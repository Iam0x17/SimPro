package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	SSH struct {
		Port       int               `yaml:"port"`
		ValidUsers map[string]string `yaml:"valid_users"`
		Commands   map[string]string `yaml:"commands"`
	} `yaml:"ssh"`
	FTP struct {
		Port           int    `yaml:"port"`
		User           string `yaml:"user"`
		Pass           string `yaml:"pass"`
		ReadOnly       bool   `yaml:"read_only"`
		WelcomeMessage string `yaml:"welcome_message"`
	} `yaml:"ftp"`
	Redis struct {
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"redis"`
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(`D:\code\go\ProtoSimService\config.yaml`)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &cfg, nil
}
