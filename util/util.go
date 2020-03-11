package util

import (
	"os"
	"path/filepath"
)

const (
	defaultDataDir = ".tomatobot"
	confName       = "conf.json"
)

func GetDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/root", defaultDataDir)
	}

	return filepath.Join(home, defaultDataDir)
}

func GetConfigPath() string {
	return filepath.Join(GetDataDir(), confName)
}
