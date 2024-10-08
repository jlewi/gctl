package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Note: The application uses viper for configuration management. Viper merges configurations from various sources
//such as files, environment variables, and command line flags. After merging, viper unmarshals the configuration into the Configuration struct, which is then used throughout the application.

const (
	ConfigFlagName = "config"
	LevelFlagName  = "level"
	appName        = "gctl"
	ConfigDir      = "." + appName
)

// Config represents the persistent configuration data for Foyle.
//
// Currently, the format of the data on disk and in memory is identical. In the future, we may modify this to simplify
// changes to the disk format and to store in-memory values that should not be written to disk.
type Config struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion" yamltags:"required"`
	Kind       string `json:"kind" yaml:"kind" yamltags:"required"`

	Logging Logging `json:"logging" yaml:"logging"`
	// OAuthClientFile is the path to the JSON file containing the OAuth client secret.
	OAuthClientFile string `json:"oauthClientFile,omitempty" yaml:"oauthClientFile,omitempty"`
}

type Logging struct {
	Level  string `json:"level,omitempty" yaml:"level,omitempty"`
	LogDir string `json:"logDir,omitempty" yaml:"logDir,omitempty"`
}

func (c *Config) GetLogLevel() string {
	if c.Logging.Level == "" {
		return "info"
	}
	return c.Logging.Level
}

func (c *Config) GetLogDir() string {
	if c.Logging.LogDir == "" {
		return os.TempDir()
	}
	return c.Logging.LogDir
}

// GetConfigDir returns the configuration directory
func (c *Config) GetConfigDir() string {
	configFile := viper.ConfigFileUsed()
	if configFile != "" {
		return filepath.Dir(configFile)
	}

	// Since there is no config file we will use the default config directory.
	return binHome()
}

// GetOAuthCredentialsFile returns the path to the file where the refresh token should be stored.
func (c *Config) GetOAuthCredentialsFile() string {
	return filepath.Join(c.GetConfigDir(), "credentials.json")
}

// IsValid validates the configuration and returns any errors.
func (c *Config) IsValid() []string {
	problems := make([]string, 0, 1)
	return problems
}

// DeepCopy returns a deep copy.
func (c *Config) DeepCopy() Config {
	b, err := json.Marshal(c)
	if err != nil {
		log := zapr.NewLogger(zap.L())
		log.Error(err, "Failed to marshal config")
		panic(err)
	}
	var copy Config
	if err := json.Unmarshal(b, &copy); err != nil {
		log := zapr.NewLogger(zap.L())
		log.Error(err, "Failed to unmarshal config")
		panic(err)
	}
	return copy
}

// InitViper function is responsible for reading the configuration file and environment variables, if they are set.
// The results are stored in viper. To retrieve a configuration, use the GetConfig function.
// The function accepts a cmd parameter which allows binding to command flags.
func InitViper(cmd *cobra.Command) error {
	// Ref https://github.com/spf13/viper#establishing-defaults
	viper.SetEnvPrefix(appName)
	// name of config file (without extension)
	viper.SetConfigName("config")
	// make home directory the first search path
	viper.AddConfigPath("$HOME/." + appName)

	// Without the replacer overriding with environment variables doesn't work
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// We need to attach to the command line flag if it was specified.
	keyToflagName := map[string]string{
		ConfigFlagName:             ConfigFlagName,
		"logging." + LevelFlagName: LevelFlagName,
	}

	if cmd != nil {
		for key, flag := range keyToflagName {
			if err := viper.BindPFlag(key, cmd.Flags().Lookup(flag)); err != nil {
				return err
			}
		}
	}

	// Ensure the path for the config file path is set
	// Required since we use viper to persist the location of the config file so can save to it.
	cfgFile := viper.GetString(ConfigFlagName)
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log := zapr.NewLogger(zap.L())
			log.Error(err, "config file not found", "file", cfgFile)
			return nil
		}
		if _, ok := err.(*fs.PathError); ok {
			log := zapr.NewLogger(zap.L())
			log.Error(err, "config file not found", "file", cfgFile)
			return nil
		}
		return err
	}
	return nil
}

// GetConfig returns a configuration created from the viper configuration.
func GetConfig() *Config {
	// We do this as a way to load the configuration while still allowing values to be overwritten by viper
	cfg := &Config{}

	if err := viper.Unmarshal(cfg); err != nil {
		panic(fmt.Errorf("failed to unmarshal configuration; error %v", err))
	}

	return cfg
}

func binHome() string {
	log := zapr.NewLogger(zap.L())
	usr, err := user.Current()
	homeDir := ""
	if err != nil {
		log.Error(err, "failed to get current user; falling back to temporary directory for homeDir", "homeDir", os.TempDir())
		homeDir = os.TempDir()
	} else {
		homeDir = usr.HomeDir
	}
	p := filepath.Join(homeDir, ConfigDir)

	return p
}

// Write saves the configuration to a file.
func (c *Config) Write(cfgFile string) error {
	log := zapr.NewLogger(zap.L())
	if cfgFile == "" {
		return errors.Errorf("no config file specified")
	}
	configDir := filepath.Dir(cfgFile)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		log.Info("creating config directory", "dir", configDir)
		if err := os.Mkdir(configDir, 0700); err != nil {
			return errors.Wrapf(err, "Ffailed to create config directory %s", configDir)
		}
	}

	f, err := os.Create(cfgFile)
	if err != nil {
		return err
	}

	return yaml.NewEncoder(f).Encode(c)
}

func DefaultConfigFile() string {
	return binHome() + "/config.yaml"
}
