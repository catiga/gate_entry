package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

type LogConfig struct {
	Path  string `yaml:"path"`
	Level string `yaml:"level"`
}

type ProxyConfig struct {
	ServerPort int    `yaml:"serverPort"`
	TargetURI  string `yaml:"targetURI"`
	Code       string `yaml:"code"`
}

type Config struct {
	Log   LogConfig   `yaml:"log"`
	Proxy ProxyConfig `yaml:"proxy"`
}

var config Config

func findProjectRoot(currentDir, rootIndicator string) (string, error) {
	if _, err := os.Stat(filepath.Join(currentDir, rootIndicator)); err == nil {
		return currentDir, nil
	}
	parentDir := filepath.Dir(currentDir)
	if currentDir == parentDir {
		return "", os.ErrNotExist
	}
	return findProjectRoot(parentDir, rootIndicator)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func init() {
	var confFilePath string

	if configFilePathFromEnv := os.Getenv("DALINK_GO_CONFIG_PATH"); configFilePathFromEnv != "" {
		confFilePath = configFilePathFromEnv
	} else {
		_, filename, _, _ := runtime.Caller(0)
		testDir := filepath.Dir(filename)
		confFilePath, _ = findProjectRoot(testDir, "__mark__")
		if len(confFilePath) > 0 {
			confFilePath += "/config/conf.yml"
		}
	}
	if len(confFilePath) == 0 {
		// find in current directory
		exePath, _ := os.Executable()
		exeDir := filepath.Dir(exePath)
		confFilePath = filepath.Join(exeDir, "conf.yml")
		if !fileExists(confFilePath) {
			log.Fatal("System root directory setting error.")
		}
	}
	log.Println("current config file ", confFilePath)

	viper.SetConfigFile(confFilePath)

	// viper.SetConfigType("yml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Unable to read configuration file: %s", err)
	}
	targetStr := viper.AllSettings()
	fmt.Printf("Raw config: %+v\n", targetStr)
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to parse configuration: %s", err)
	}
}

func Get() *Config {
	return &config
}
