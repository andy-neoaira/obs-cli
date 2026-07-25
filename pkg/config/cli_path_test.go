package config_test

import (
	"errors"
	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestConfigCliPath(t *testing.T) {
	originalUserConfigDirectory := config.UserConfigDirectory
	defer func() { config.UserConfigDirectory = originalUserConfigDirectory }()

	t.Run("UserConfigDir func returns a directory", func(t *testing.T) {
		// Arrange
		config.UserConfigDirectory = func() (string, error) {
			return "user/config/dir", nil
		}
		// Act
		obsConfigDir, obsConfigFile, err := config.CliPath()
		// Assert
		assert.Equal(t, nil, err)
		assert.Equal(t, "user/config/dir/obs-cli", obsConfigDir)
		assert.Equal(t, "user/config/dir/obs-cli/preferences.json", obsConfigFile)
	})

	t.Run("UserConfigDir func returns an error", func(t *testing.T) {
		// Arrange
		config.UserConfigDirectory = func() (string, error) {
			return "", errors.New(config.UserConfigDirectoryNotFoundErrorMessage)
		}
		// Act
		obsConfigDir, obsConfigFile, err := config.CliPath()
		// Assert
		assert.Equal(t, config.UserConfigDirectoryNotFoundErrorMessage, err.Error())
		assert.Equal(t, "", obsConfigDir)
		assert.Equal(t, "", obsConfigFile)
	})

}

func TestConfigCliPathEnvironmentOverride(t *testing.T) {
	originalUserConfigDirectory := config.UserConfigDirectory
	t.Cleanup(func() { config.UserConfigDirectory = originalUserConfigDirectory })
	config.UserConfigDirectory = originalUserConfigDirectory

	t.Run("uses an isolated absolute config root", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "config-root")
		t.Setenv("OBS_CLI_CONFIG_HOME", root)

		configDir, configFile, err := config.V2Path()

		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(root, "obs-cli"), configDir)
		assert.Equal(t, filepath.Join(root, "obs-cli", "config-v2.json"), configFile)
	})

	t.Run("rejects a relative override", func(t *testing.T) {
		t.Setenv("OBS_CLI_CONFIG_HOME", "relative/config-root")

		configDir, configFile, err := config.V2Path()

		assert.EqualError(t, err, config.UserConfigDirectoryNotFoundErrorMessage)
		assert.Empty(t, configDir)
		assert.Empty(t, configFile)
	})
}
