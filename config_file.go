package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

const HomeEnvVar = "PRETTY_LOGRUS_HOME"
const configFileName = "config.json"

func hasHomeEnvVar() bool {
	_, ok := os.LookupEnv(HomeEnvVar)
	return ok
}

func hasConfigFile() bool {
	env, ok := os.LookupEnv(HomeEnvVar)
	if !ok {
		return false
	}

	_, err := os.Stat(path.Join(env, configFileName))
	return !os.IsNotExist(err)
}

func mkdirp(path string) error {
	return os.MkdirAll(path, 0755)
}

func pathToHomeEnvDirExists() (bool, error) {
	dir := os.Getenv(HomeEnvVar)
	fullPath := path.Join(dir, configFileName)

	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ensureConfigFileExistsIfHomeEnvIsSet(defaultConfig *Config) error {
	// If PRETTY_LOGRUS_HOME is not set, don't do anything
	if !hasHomeEnvVar() {
		return nil
	}

	// If a config file already exists, don't do anything'
	if hasConfigFile() {
		return nil
	}

	// If PRETTY_LOGRUS_HOME is set, but the directory structure/path doesn't exist, create it
	if exists, err := pathToHomeEnvDirExists(); !exists || err != nil {
		if err := mkdirp(os.Getenv(HomeEnvVar)); err != nil {
			return err
		}
	}

	// Create the config file
	if err := writeConfigFile(defaultConfig); err != nil {
		return err
	}

	return nil
}

func writeConfigFile(configFile *Config) error {
	content, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(os.Getenv(HomeEnvVar), configFileName), content, 0644)
}

func readConfigFile(defaultConfig Config) (*Config, error) {
	homeDir := os.Getenv(HomeEnvVar)
	configFilePath := path.Join(homeDir, configFileName)

	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	config := defaultConfig
	if err = json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err)
	}

	if isDebug() {
		configJson, _ := json.MarshalIndent(config, "", "  ")
		fmt.Printf("Read config file: %s\n", configJson)
	}

	return &config, nil
}
