package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	configFileName = ".gatorconfig.json"
)

type Config struct {
	Username string `json:"current_user_name"`
	DBUrl    string `json:"db_url"`
}

func Read() Config {
	filePath, err := getConfigFilePath()
	if err != nil {
		fmt.Println("Config file not found at ~/.gatorconfig.json")
		os.Exit(1)
	}
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Unable to open '" + filePath + "'")
		os.Exit(1)
	}
	defer f.Close()

	byteValues, _ := io.ReadAll(f)
	var cfg Config
	if err = json.Unmarshal(byteValues, &cfg); err != nil {
		fmt.Println("Unable to unmarshal ~/.gatorconfig.json")
		os.Exit(1)
	}
	return cfg
}

func (config *Config) SetUser(user string) error {
	config.Username = user
	return write(*config)
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home + "/" + configFileName, nil
}

func write(config Config) error {

	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	jsonString, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, jsonString, os.ModePerm)
	return err
}
