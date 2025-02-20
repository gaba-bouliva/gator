package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func Read() (*Config, error) {
	filePath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	jsBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var newConfig Config

	err = json.Unmarshal(jsBytes, &newConfig)
	if err != nil {
		return nil, err
	}

	return &newConfig, nil

}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username

	oldConfig, err := Read()
	if err != nil {
		return err
	}
	oldConfig.CurrentUserName = username
	err = wirte(*oldConfig)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) GetCurrentUser() (string, error) {
	oldConfig, err := Read()
	if err != nil {
		return "", err
	}

	return oldConfig.CurrentUserName, nil
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(homeDir, configFileName)

	return filePath, nil
}

func wirte(cfg Config) error {
	updatedConfig, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(updatedConfig)
	if err != nil {
		fmt.Println("error writing to file")
		return err
	}

	return nil
}
