package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const ConfigFileName = "plr-config.json"

type MsgIgnoreEntry struct {
	Text string `json:"text"`
	App  string `json:"app"`
}

type AppConfigFile struct {
	MessageIgnoreList []MsgIgnoreEntry `json:"message_ignore_list"`
}

func saveAppConfigFile(configFile AppConfigFile) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	filePath := filepath.Join(configDir, "pretty-logrus", ConfigFileName)

	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print JSON
	if err := encoder.Encode(configFile); err != nil {
		return fmt.Errorf("could not write JSON to file: %w", err)
	}

	if isDebug() {
		fmt.Printf("app config file written to: %s\n", filePath)
	}

	return nil
}

func readAppConfigFile() (*AppConfigFile, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	filePath := filepath.Join(configDir, "pretty-logrus", ConfigFileName)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	var configFile AppConfigFile
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configFile); err != nil {
		return nil, fmt.Errorf("could not read JSON from file: %w", err)
	}

	if isDebug() {
		fmt.Printf("app config file read from: %s\n", filePath)
	}

	return &configFile, nil
}
