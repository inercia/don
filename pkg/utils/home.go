// Package utils provides utility functions for Don
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// DonDirEnv is the environment variable that specifies the configuration directory for Don
	DonDirEnv = "DON_DIR"
	// DonConfigEnv is the environment variable that specifies the agent configuration file path
	DonConfigEnv = "DON_CONFIG"
	// DonHome is the name of the configuration directory for Don
	DonHome = ".don"
)

// GetHome returns the user's home directory in a portable way
func GetHome() (string, error) {
	var home string

	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
		if home == "" {
			home = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		}
	} else {
		home = os.Getenv("HOME")
	}

	if home == "" {
		return "", fmt.Errorf("unable to determine home directory")
	}

	return home, nil
}

// GetDonHome returns the Don configuration directory
// This is typically ~/.don on Unix-like systems or %USERPROFILE%\.don on Windows
func GetDonHome() (string, error) {
	if donHome := os.Getenv(DonDirEnv); donHome != "" {
		return donHome, nil
	}

	home, err := GetHome()
	if err != nil {
		return "", err
	}

	donHome := filepath.Join(home, DonHome)
	return donHome, nil
}
