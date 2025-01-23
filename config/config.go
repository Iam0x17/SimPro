package config

import (
	"embed"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

var (
	AssetsFs      embed.FS
	SshPrivateKey = "assets/ssh_private.key"
	ConfigFile    = "assets/config.yaml"
)

type Config struct {
	SSH struct {
		Port     string            `yaml:"port"`
		User     string            `yaml:"user"`
		Pass     string            `yaml:"pass"`
		Commands map[string]string `yaml:"commands"`
	} `yaml:"ssh"`
	FTP struct {
		Port string `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
	} `yaml:"ftp"`
	Telnet struct {
		Port string `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
	} `yaml:"telnet"`
	Redis struct {
		Port string `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
	} `yaml:"redis"`
	Postgres struct {
		Port string `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
	} `yaml:"postgres"`
	MySql struct {
		Port string `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
	} `yaml:"mysql"`
}

func LoadConfig(path string) (*Config, error) {
	var data []byte
	var err error

	if path != "" {
		data, err = os.ReadFile(path)

	} else {
		data, err = AssetsFs.ReadFile(ConfigFile)
	}
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

func GetEmbed(assets embed.FS) {
	AssetsFs = assets
}
