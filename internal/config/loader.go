// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-viper/encoding/ini"
	"github.com/itential/ipctl/internal/app"
	"github.com/itential/ipctl/internal/profile"
	"github.com/itential/ipctl/internal/repository"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Loader provides a fluent API for loading configuration from multiple sources.
// It separates the concerns of file loading, environment variable binding, and
// flag parsing, making the configuration loading process testable and maintainable.
//
// The loader follows this precedence order:
//  1. Defaults (lowest priority)
//  2. Configuration file
//  3. Environment variables
//  4. Command-line flags (highest priority)
//
// Example usage:
//
//	loader := NewLoader()
//	cfg, err := loader.
//		WithConfigFile("/path/to/config").
//		WithWorkingDir("~/.platform.d").
//		Load()
//	if err != nil {
//		// handle error
//	}
type Loader struct {
	// configFile is the explicit path to a configuration file.
	// If empty, the loader will search in standard locations.
	configFile string

	// workingDir is the user's configuration directory.
	// Defaults to ~/.platform.d
	workingDir string

	// sysConfigPath is the system-wide configuration directory.
	// Defaults to /etc/ipctl
	sysConfigPath string

	// profileFlag is the profile name specified via command-line flag.
	profileFlag string

	// configFileFlag is the config file path specified via command-line flag.
	configFileFlag string

	// fileName is the base name of the configuration file to search for.
	// Defaults to "config"
	fileName string

	// defaults contains the default values for all configuration keys.
	defaults map[string]interface{}

	// envBindings maps configuration keys to environment variable names.
	envBindings map[string]string
}

// NewLoader creates a new configuration loader with sensible defaults.
// The returned loader can be customized using the With* methods before
// calling Load().
func NewLoader() *Loader {
	return &Loader{
		workingDir:    "~/.platform.d",
		sysConfigPath: "/etc/ipctl",
		fileName:      defaultFileName,
		defaults:      defaultValues,
		envBindings:   defaultEnvVarBindings,
	}
}

// WithConfigFile sets an explicit path to a configuration file.
// This takes precedence over the standard search paths.
func (l *Loader) WithConfigFile(path string) *Loader {
	l.configFile = path
	return l
}

// WithWorkingDir sets the user's configuration directory.
// This is typically ~/.platform.d and will be searched for configuration files.
func (l *Loader) WithWorkingDir(dir string) *Loader {
	l.workingDir = dir
	return l
}

// WithSysConfigPath sets the system-wide configuration directory.
// This is typically /etc/ipctl and will be searched for configuration files.
func (l *Loader) WithSysConfigPath(path string) *Loader {
	l.sysConfigPath = path
	return l
}

// WithFileName sets the base name of the configuration file to search for.
// Defaults to "config".
func (l *Loader) WithFileName(name string) *Loader {
	l.fileName = name
	return l
}

// WithDefaults sets the default values for configuration keys.
// If not set, uses the package-level defaultValues.
func (l *Loader) WithDefaults(defaults map[string]interface{}) *Loader {
	l.defaults = defaults
	return l
}

// WithEnvBindings sets the environment variable bindings.
// If not set, uses the package-level defaultEnvVarBindings.
func (l *Loader) WithEnvBindings(bindings map[string]string) *Loader {
	l.envBindings = bindings
	return l
}

// Load loads configuration from all sources and returns a Config instance.
// It returns an error if any step of the loading process fails.
//
// The loading process:
//  1. Parse command-line flags (--config, --profile)
//  2. Create a new viper instance
//  3. Apply defaults
//  4. Bind environment variables
//  5. Load configuration file
//  6. Build Config struct
//  7. Load profiles and repositories
func (l *Loader) Load() (*Config, error) {
	// Parse command-line flags first, as they affect subsequent steps
	if err := l.parseFlags(); err != nil {
		return nil, fmt.Errorf("parsing flags: %w", err)
	}

	// Create a new viper instance (not global) with INI support restored.
	// Viper 1.20 removed the built-in INI codec, so it must be registered
	// explicitly; the registry falls back to the built-in YAML, JSON, and
	// TOML codecs for other formats.
	registry := viper.NewCodecRegistry()
	if err := registry.RegisterCodec("ini", ini.Codec{}); err != nil {
		return nil, fmt.Errorf("registering ini codec: %w", err)
	}
	v := viper.NewWithOptions(viper.WithCodecRegistry(registry))

	// Apply defaults
	l.applyDefaults(v)

	// Bind environment variables
	if err := l.bindEnvVars(v); err != nil {
		return nil, fmt.Errorf("binding environment variables: %w", err)
	}

	// Load configuration file
	if err := l.loadConfigFile(v); err != nil {
		return nil, fmt.Errorf("loading config file: %w", err)
	}

	// Build the Config struct
	cfg, err := l.buildConfig(v)
	if err != nil {
		return nil, fmt.Errorf("building config: %w", err)
	}

	return cfg, nil
}

// parseFlags parses command-line flags to extract --config and --profile.
// These flags affect how the configuration is loaded.
func (l *Loader) parseFlags() error {
	flagSet := pflag.NewFlagSet("config", pflag.ContinueOnError)
	flagSet.StringVar(&l.configFileFlag, "config", "", "Path to config file")
	flagSet.StringVar(&l.profileFlag, "profile", "", "Connection profile")
	flagSet.ParseErrorsWhitelist.UnknownFlags = true // Ignore unknown flags
	flagSet.Usage = func() {}                        // Suppress default usage message

	if err := flagSet.Parse(os.Args[1:]); err != nil && err != pflag.ErrHelp {
		return fmt.Errorf("parsing command line arguments: %w", err)
	}

	// Environment variables can override flags
	if envConfig := os.Getenv("IPCTL_CONFIG_FILE"); envConfig != "" {
		l.configFileFlag = envConfig
	}

	// Validate and expand the config file path if provided
	if l.configFileFlag != "" {
		if _, err := os.Stat(l.configFileFlag); os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", l.configFileFlag)
		}

		expanded, err := homedir.Expand(l.configFileFlag)
		if err != nil {
			return fmt.Errorf("expanding config file path: %w", err)
		}
		l.configFileFlag = expanded
	}

	return nil
}

// applyDefaults sets all default values in the viper instance.
func (l *Loader) applyDefaults(v *viper.Viper) {
	for key, value := range l.defaults {
		v.SetDefault(key, value)
	}
}

// bindEnvVars binds environment variables to configuration keys.
func (l *Loader) bindEnvVars(v *viper.Viper) error {
	for key, envVar := range l.envBindings {
		if err := v.BindEnv(key, envVar); err != nil {
			return fmt.Errorf("binding environment variable %s to key %s: %w", envVar, key, err)
		}
	}
	return nil
}

// loadConfigFile loads the configuration file from the appropriate location.
// It searches in the following order:
//  1. Explicit configFile set via WithConfigFile
//  2. Path from --config flag
//  3. Path from IPCTL_CONFIG environment variable
//  4. Standard search paths (workingDir, sysConfigPath)
//
// The function automatically detects the file format based on the file extension.
// Supported formats:
//   - .ini (default)
//   - .yaml, .yml (YAML)
//   - .toml (TOML)
//   - .json (JSON)
//
// If no extension is provided when searching standard paths, the loader will
// try each format in order: INI, YAML, TOML, JSON.
func (l *Loader) loadConfigFile(v *viper.Viper) error {
	var explicitFile string

	// Priority 1: Explicit config file from WithConfigFile
	if l.configFile != "" {
		explicitFile = l.configFile
	}

	// Priority 2: Config file from --config flag
	if l.configFileFlag != "" {
		explicitFile = l.configFileFlag
	}

	// Priority 3: Config file from IPCTL_CONFIG environment variable
	if envConfig := os.Getenv("IPCTL_CONFIG"); envConfig != "" {
		explicitFile = envConfig
	}

	// If an explicit file was specified, detect format and load it
	if explicitFile != "" {
		configType := detectConfigType(explicitFile)
		v.SetConfigType(configType)
		v.SetConfigFile(explicitFile)
	} else {
		// No explicit file - search in standard locations
		// Default to INI for backward compatibility
		v.SetConfigType("ini")
		v.SetConfigName(l.fileName)

		// Add working directory to search path
		expandedWorkingDir, err := homedir.Expand(l.workingDir)
		if err != nil {
			return fmt.Errorf("expanding working directory %s: %w", l.workingDir, err)
		}
		v.AddConfigPath(expandedWorkingDir)

		// Add system config path to search path
		v.AddConfigPath(l.sysConfigPath)
	}

	// Try to read the config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("reading config file: %w", err)
		}
	}

	return nil
}

// detectConfigType determines the configuration file format based on the file extension.
// Returns the format type that Viper understands.
//
// Supported formats:
//   - .ini -> "ini"
//   - .yaml, .yml -> "yaml"
//   - .toml -> "toml"
//   - .json -> "json"
//
// Defaults to "ini" if the extension is not recognized or missing.
func detectConfigType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".json":
		return "json"
	case ".ini":
		return "ini"
	default:
		// Default to INI for backward compatibility
		return "ini"
	}
}

// buildConfig constructs a Config instance from the loaded configuration.
func (l *Loader) buildConfig(v *viper.Viper) (*Config, error) {
	cfg := &Config{
		Settings: &app.Settings{},
	}

	// Populate application settings
	workingDir, err := homedir.Expand(v.GetString("application.working_dir"))
	if err != nil {
		return nil, fmt.Errorf("expanding working directory: %w", err)
	}
	cfg.Settings.WorkingDir = workingDir
	cfg.Settings.DefaultProfile = v.GetString("application.default_profile")
	cfg.Settings.DefaultRepository = v.GetString("application.default_repository")

	// Populate features
	cfg.Settings.Features.DatasetsEnabled = v.GetBool("features.datasets_enabled")

	// Populate git config
	cfg.Settings.Git.Name = v.GetString("git.name")
	cfg.Settings.Git.Email = v.GetString("git.email")
	cfg.Settings.Git.User = v.GetString("git.user")

	// Initialize managers
	cfg.profileManager = profile.NewManager()
	cfg.repositoryManager = repository.NewManager()

	// Load profiles
	if err := l.loadProfiles(cfg, v); err != nil {
		return nil, fmt.Errorf("loading profiles: %w", err)
	}

	// Load repositories
	if err := l.loadRepositories(cfg, v); err != nil {
		return nil, fmt.Errorf("loading repositories: %w", err)
	}

	// Set the active profile
	activeProfile := l.profileFlag
	if activeProfile == "" {
		activeProfile = cfg.Settings.DefaultProfile
	}
	if activeProfile != "" {
		cfg.profileManager.SetActive(activeProfile)
	}

	return cfg, nil
}

// loadProfiles loads all profiles from the configuration.
// It loads the default profile first, then loads all named profiles.
func (l *Loader) loadProfiles(cfg *Config, v *viper.Viper) error {
	// Load the default profile
	var defaults map[string]interface{}
	if value, exists := v.AllSettings()["profile default"]; exists {
		defaults = value.(map[string]interface{})
	} else {
		defaults = map[string]interface{}{}
	}

	// Create and add default profile
	loader := profile.NewLoader(defaults, defaults, map[string]interface{}{})
	cfg.profileManager.Add("default", loader.Load())

	// Load named profiles
	for key, value := range v.AllSettings() {
		if !strings.HasPrefix(key, "profile ") {
			continue
		}

		parts := strings.Split(key, " ")
		if len(parts) > 2 {
			return fmt.Errorf("profile name cannot contain spaces: %s", key)
		}

		profileName := parts[1]
		if profileName == "default" {
			continue // Already loaded above
		}

		values, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		// Check for environment variable overrides
		overrides := l.getProfileOverrides(profileName)

		// Create and add profile
		loader := profile.NewLoader(values, defaults, overrides)
		cfg.profileManager.Add(profileName, loader.Load())
	}

	return nil
}

// loadRepositories loads all repositories from the configuration.
func (l *Loader) loadRepositories(cfg *Config, v *viper.Viper) error {
	for key, value := range v.AllSettings() {
		if !strings.HasPrefix(key, "repository ") {
			continue
		}

		parts := strings.Split(key, " ")
		if len(parts) > 2 {
			return fmt.Errorf("repository name cannot contain spaces: %s", key)
		}

		repoName := parts[1]

		values, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		// Check for environment variable overrides
		overrides := l.getRepositoryOverrides(repoName)

		// Create and add repository
		loader := repository.NewLoader(values, overrides)
		cfg.repositoryManager.Add(repoName, loader.Load())
	}

	return nil
}

// getProfileOverrides retrieves environment variable overrides for a profile.
func (l *Loader) getProfileOverrides(name string) map[string]interface{} {
	overrides := map[string]interface{}{}
	fields := []string{"host", "port", "use_tls", "verify", "username", "password", "client_id", "client_secret", "mongo_url", "timeout"}

	for _, field := range fields {
		envKey := fmt.Sprintf("IPCTL_PROFILE_%s_%s", strings.ToUpper(name), strings.ToUpper(field))
		if val, exists := os.LookupEnv(envKey); exists {
			overrides[field] = val
		}
	}

	return overrides
}

// getRepositoryOverrides retrieves environment variable overrides for a repository.
func (l *Loader) getRepositoryOverrides(name string) map[string]interface{} {
	overrides := map[string]interface{}{}
	fields := []string{"url", "private_key", "private_key_file", "reference"}

	for _, field := range fields {
		envKey := fmt.Sprintf("IPCTL_REPOSITORY_%s_%s", strings.ToUpper(name), strings.ToUpper(field))
		if val, exists := os.LookupEnv(envKey); exists {
			overrides[field] = val
		}
	}

	return overrides
}
