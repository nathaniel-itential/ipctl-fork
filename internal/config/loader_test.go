// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/itential/ipctl/internal/app"
)

// TestNewLoader verifies that NewLoader creates a loader with proper defaults.
func TestNewLoader(t *testing.T) {
	loader := NewLoader()

	if loader == nil {
		t.Fatal("NewLoader() returned nil")
	}

	if loader.workingDir != "~/.platform.d" {
		t.Errorf("workingDir = %q, want %q", loader.workingDir, "~/.platform.d")
	}

	if loader.sysConfigPath != "/etc/ipctl" {
		t.Errorf("sysConfigPath = %q, want %q", loader.sysConfigPath, "/etc/ipctl")
	}

	if loader.fileName != defaultFileName {
		t.Errorf("fileName = %q, want %q", loader.fileName, defaultFileName)
	}

	if loader.defaults == nil {
		t.Error("defaults should not be nil")
	}

	if loader.envBindings == nil {
		t.Error("envBindings should not be nil")
	}
}

// TestLoaderWithConfigFile verifies that WithConfigFile sets the config file path.
func TestLoaderWithConfigFile(t *testing.T) {
	loader := NewLoader().WithConfigFile("/custom/config.ini")

	if loader.configFile != "/custom/config.ini" {
		t.Errorf("configFile = %q, want %q", loader.configFile, "/custom/config.ini")
	}
}

// TestLoaderWithWorkingDir verifies that WithWorkingDir sets the working directory.
func TestLoaderWithWorkingDir(t *testing.T) {
	loader := NewLoader().WithWorkingDir("/custom/workdir")

	if loader.workingDir != "/custom/workdir" {
		t.Errorf("workingDir = %q, want %q", loader.workingDir, "/custom/workdir")
	}
}

// TestLoaderWithSysConfigPath verifies that WithSysConfigPath sets the system config path.
func TestLoaderWithSysConfigPath(t *testing.T) {
	loader := NewLoader().WithSysConfigPath("/custom/etc")

	if loader.sysConfigPath != "/custom/etc" {
		t.Errorf("sysConfigPath = %q, want %q", loader.sysConfigPath, "/custom/etc")
	}
}

// TestLoaderWithFileName verifies that WithFileName sets the config file name.
func TestLoaderWithFileName(t *testing.T) {
	loader := NewLoader().WithFileName("custom-config")

	if loader.fileName != "custom-config" {
		t.Errorf("fileName = %q, want %q", loader.fileName, "custom-config")
	}
}

// TestLoaderWithDefaults verifies that WithDefaults sets custom defaults.
func TestLoaderWithDefaults(t *testing.T) {
	customDefaults := map[string]interface{}{
		"test.key": "test-value",
	}

	loader := NewLoader().WithDefaults(customDefaults)

	if loader.defaults["test.key"] != "test-value" {
		t.Errorf("defaults[test.key] = %v, want %v", loader.defaults["test.key"], "test-value")
	}
}

// TestLoaderWithEnvBindings verifies that WithEnvBindings sets custom environment bindings.
func TestLoaderWithEnvBindings(t *testing.T) {
	customBindings := map[string]string{
		"test.key": "TEST_ENV",
	}

	loader := NewLoader().WithEnvBindings(customBindings)

	if loader.envBindings["test.key"] != "TEST_ENV" {
		t.Errorf("envBindings[test.key] = %q, want %q", loader.envBindings["test.key"], "TEST_ENV")
	}
}

// TestLoaderFluentAPI verifies that all With* methods return the loader for chaining.
func TestLoaderFluentAPI(t *testing.T) {
	loader := NewLoader().
		WithConfigFile("/test/config").
		WithWorkingDir("/test/work").
		WithSysConfigPath("/test/etc").
		WithFileName("test-config").
		WithDefaults(map[string]interface{}{"key": "value"}).
		WithEnvBindings(map[string]string{"key": "ENV"})

	if loader.configFile != "/test/config" {
		t.Errorf("configFile = %q, want %q", loader.configFile, "/test/config")
	}
	if loader.workingDir != "/test/work" {
		t.Errorf("workingDir = %q, want %q", loader.workingDir, "/test/work")
	}
	if loader.sysConfigPath != "/test/etc" {
		t.Errorf("sysConfigPath = %q, want %q", loader.sysConfigPath, "/test/etc")
	}
	if loader.fileName != "test-config" {
		t.Errorf("fileName = %q, want %q", loader.fileName, "test-config")
	}
}

// TestLoaderLoadWithDefaultsOnly verifies that Load works with defaults only (no config file).
func TestLoaderLoadWithDefaultsOnly(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl"}

	loader := NewLoader().WithWorkingDir("/nonexistent")

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Verify defaults were applied
	defaults := app.DefaultValues()
	expectedGitUser := defaults["git.user"].(string)
	expectedDatasetsEnabled := defaults["features.datasets_enabled"].(bool)

	if cfg.Settings.Git.User != expectedGitUser {
		t.Errorf("cfg.Settings.Git.User = %q, want %q", cfg.Settings.Git.User, expectedGitUser)
	}

	if cfg.Settings.Features.DatasetsEnabled != expectedDatasetsEnabled {
		t.Errorf("cfg.Settings.Features.DatasetsEnabled = %v, want %v", cfg.Settings.Features.DatasetsEnabled, expectedDatasetsEnabled)
	}
}

// TestLoaderLoadWithConfigFile verifies that Load reads configuration from a file.
func TestLoaderLoadWithConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config")

	configContent := `[application]
working_dir = /test/working
default_profile = testprofile
default_repository = testrepo

[features]
datasets_enabled = true

[git]
name = Test User
email = test@example.com
user = testgit

[profile default]
host = default.example.com
port = 8080
use_tls = true
verify = false
username = defaultuser
password = defaultpass

[profile production]
host = prod.example.com
port = 443
use_tls = true
verify = true
username = produser

[repository myrepo]
url = https://github.com/example/repo.git
reference = main
`

	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl"}

	loader := NewLoader().WithConfigFile(configFile)

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Verify application settings
	if cfg.Settings.WorkingDir != "/test/working" {
		t.Errorf("cfg.Settings.WorkingDir = %q, want %q", cfg.Settings.WorkingDir, "/test/working")
	}

	if cfg.Settings.DefaultProfile != "testprofile" {
		t.Errorf("cfg.Settings.DefaultProfile = %q, want %q", cfg.Settings.DefaultProfile, "testprofile")
	}

	if cfg.Settings.DefaultRepository != "testrepo" {
		t.Errorf("cfg.Settings.DefaultRepository = %q, want %q", cfg.Settings.DefaultRepository, "testrepo")
	}

	// Verify features
	if !cfg.Settings.Features.DatasetsEnabled {
		t.Error("cfg.Settings.Features.DatasetsEnabled = false, want true")
	}

	// Verify git config
	if cfg.Settings.Git.Name != "Test User" {
		t.Errorf("cfg.Settings.Git.Name = %q, want %q", cfg.Settings.Git.Name, "Test User")
	}

	if cfg.Settings.Git.Email != "test@example.com" {
		t.Errorf("cfg.Settings.Git.Email = %q, want %q", cfg.Settings.Git.Email, "test@example.com")
	}

	if cfg.Settings.Git.User != "testgit" {
		t.Errorf("cfg.Settings.Git.User = %q, want %q", cfg.Settings.Git.User, "testgit")
	}

	// Verify profiles
	defaultProfile, err := cfg.GetProfile("default")
	if err != nil {
		t.Errorf("GetProfile(default) returned error: %v", err)
	}

	if defaultProfile.Host != "default.example.com" {
		t.Errorf("defaultProfile.Host = %q, want %q", defaultProfile.Host, "default.example.com")
	}

	if defaultProfile.Port != 8080 {
		t.Errorf("defaultProfile.Port = %d, want %d", defaultProfile.Port, 8080)
	}

	if !defaultProfile.UseTLS {
		t.Error("defaultProfile.UseTLS = false, want true")
	}

	if defaultProfile.Verify {
		t.Error("defaultProfile.Verify = true, want false")
	}

	prodProfile, err := cfg.GetProfile("production")
	if err != nil {
		t.Errorf("GetProfile(production) returned error: %v", err)
	}

	if prodProfile.Host != "prod.example.com" {
		t.Errorf("prodProfile.Host = %q, want %q", prodProfile.Host, "prod.example.com")
	}

	// Verify repository
	repo, err := cfg.GetRepository("myrepo")
	if err != nil {
		t.Errorf("GetRepository(myrepo) returned error: %v", err)
	}

	if repo.Url != "https://github.com/example/repo.git" {
		t.Errorf("repo.Url = %q, want %q", repo.Url, "https://github.com/example/repo.git")
	}

	if repo.Reference != "main" {
		t.Errorf("repo.Reference = %q, want %q", repo.Reference, "main")
	}
}

// TestLoaderLoadWithProfileFlag verifies that the --profile flag is respected.
func TestLoaderLoadWithProfileFlag(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config")

	configContent := `[application]
default_profile = default

[profile default]
host = default.example.com

[profile custom]
host = custom.example.com
`

	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl", "--profile", "custom"}

	loader := NewLoader().WithConfigFile(configFile)

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Verify the active profile is custom
	activeProfile, err := cfg.ActiveProfile()
	if err != nil {
		t.Fatalf("ActiveProfile() returned error: %v", err)
	}

	if activeProfile.Host != "custom.example.com" {
		t.Errorf("activeProfile.Host = %q, want %q", activeProfile.Host, "custom.example.com")
	}
}

// TestLoaderLoadWithEnvVars verifies that environment variables override config file values.
func TestLoaderLoadWithEnvVars(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config")

	configContent := `[application]
working_dir = /original/path

[git]
name = Original Name
`

	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variables
	os.Setenv("IPCTL_APPLICATION_WORKING_DIR", "/override/path")
	os.Setenv("IPCTL_GIT_NAME", "Override Name")
	defer func() {
		os.Unsetenv("IPCTL_APPLICATION_WORKING_DIR")
		os.Unsetenv("IPCTL_GIT_NAME")
	}()

	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl"}

	loader := NewLoader().WithConfigFile(configFile)

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Verify environment variables override config file
	if cfg.Settings.WorkingDir != "/override/path" {
		t.Errorf("cfg.Settings.WorkingDir = %q, want %q", cfg.Settings.WorkingDir, "/override/path")
	}

	if cfg.Settings.Git.Name != "Override Name" {
		t.Errorf("cfg.Settings.Git.Name = %q, want %q", cfg.Settings.Git.Name, "Override Name")
	}
}

// TestLoaderLoadWithInvalidConfigFile verifies that Load returns an error for invalid config files.
func TestLoaderLoadWithInvalidConfigFile(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl", "--config", "/nonexistent/config"}

	loader := NewLoader()

	_, err := loader.Load()
	if err == nil {
		t.Error("Load() should return error for nonexistent config file")
	}
}

// TestLoaderLoadWithMalformedConfigFile verifies that Load returns an error for malformed config files.
func TestLoaderLoadWithMalformedConfigFile(t *testing.T) {
	// Create a temporary malformed config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config")

	malformedContent := `[application
this is not valid INI
`

	if err := os.WriteFile(configFile, []byte(malformedContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl"}

	loader := NewLoader().WithConfigFile(configFile)

	_, err := loader.Load()
	if err == nil {
		t.Error("Load() should return error for malformed config file")
	}
}

// TestLoaderLoadWithProfileNameContainingSpaces verifies that Load returns an error for invalid profile names.
func TestLoaderLoadWithProfileNameContainingSpaces(t *testing.T) {
	// Create a temporary config file with invalid profile name
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config")

	configContent := `[profile my profile]
host = example.com
`

	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl"}

	loader := NewLoader().WithConfigFile(configFile)

	_, err := loader.Load()
	if err == nil {
		t.Error("Load() should return error for profile name with spaces")
	}
}

// TestLoaderLoadWithRepositoryNameContainingSpaces verifies that Load returns an error for invalid repository names.
func TestLoaderLoadWithRepositoryNameContainingSpaces(t *testing.T) {
	// Create a temporary config file with invalid repository name
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config")

	configContent := `[repository my repo]
url = https://github.com/example/repo.git
`

	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl"}

	loader := NewLoader().WithConfigFile(configFile)

	_, err := loader.Load()
	if err == nil {
		t.Error("Load() should return error for repository name with spaces")
	}
}

// TestConfigLoadingIsThreadSafe verifies that config loading can happen in parallel
// without race conditions. This ensures no global viper state is being used.
func TestConfigLoadingIsThreadSafe(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config")
	workingDir := filepath.Join(tmpDir, "working")
	configContent := fmt.Sprintf(`[application]
working_dir = %s
default_profile = default

[features]
datasets_enabled = true

[git]
name = Test User
email = test@example.com

[profile default]
host = localhost
port = 3000
use_tls = true
verify = false

[profile prod]
host = prod.example.com
port = 443
use_tls = true
verify = true

[repository myrepo]
url = https://github.com/example/repo.git
reference = main
`, workingDir)
	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl"}

	// Load config in parallel from multiple goroutines
	const numGoroutines = 20
	errCh := make(chan error, numGoroutines)
	doneCh := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			loader := NewLoader().WithConfigFile(configFile)
			cfg, err := loader.Load()
			if err != nil {
				errCh <- err
				return
			}

			// Verify the loaded config is correct
			// Note: WorkingDir gets expanded from the path in the config file
			if cfg.Settings.WorkingDir == "" {
				errCh <- fmt.Errorf("goroutine %d: Settings.WorkingDir is empty", id)
				return
			}

			if !cfg.Settings.Features.DatasetsEnabled {
				errCh <- fmt.Errorf("goroutine %d: Settings.Features.DatasetsEnabled = false, want true", id)
				return
			}

			if cfg.Settings.Git.Name != "Test User" {
				errCh <- fmt.Errorf("goroutine %d: Settings.Git.Name = %q, want Test User", id, cfg.Settings.Git.Name)
				return
			}

			// Verify profiles
			defaultProfile, err := cfg.GetProfile("default")
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d: GetProfile(default): %w", id, err)
				return
			}

			if defaultProfile.Host != "localhost" {
				errCh <- fmt.Errorf("goroutine %d: default profile host = %q, want localhost", id, defaultProfile.Host)
				return
			}

			prodProfile, err := cfg.GetProfile("prod")
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d: GetProfile(prod): %w", id, err)
				return
			}

			if prodProfile.Host != "prod.example.com" {
				errCh <- fmt.Errorf("goroutine %d: prod profile host = %q, want prod.example.com", id, prodProfile.Host)
				return
			}

			// Verify repository
			repo, err := cfg.GetRepository("myrepo")
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d: GetRepository(myrepo): %w", id, err)
				return
			}

			if repo.Url != "https://github.com/example/repo.git" {
				errCh <- fmt.Errorf("goroutine %d: repo URL = %q, want https://github.com/example/repo.git", id, repo.Url)
				return
			}

			doneCh <- struct{}{}
		}(i)
	}

	// Wait for all goroutines to complete, counting failures as
	// completions so a load error cannot deadlock the test
	completed := 0
	for completed < numGoroutines {
		select {
		case err := <-errCh:
			t.Errorf("parallel load failed: %v", err)
			completed++
		case <-doneCh:
			completed++
		}
	}
}

// TestConfigLoadingIsolation verifies that loading one config doesn't affect another.
// This ensures complete isolation between config instances.
func TestConfigLoadingIsolation(t *testing.T) {
	// Create two different config files
	tmpDir := t.TempDir()

	configFile1 := filepath.Join(tmpDir, "config1")
	configContent1 := `[application]
working_dir = /path/one
default_profile = profile1

[git]
name = User One

[profile profile1]
host = host1.example.com
port = 8001
`
	if err := os.WriteFile(configFile1, []byte(configContent1), 0600); err != nil {
		t.Fatal(err)
	}

	configFile2 := filepath.Join(tmpDir, "config2")
	configContent2 := `[application]
working_dir = /path/two
default_profile = profile2

[git]
name = User Two

[profile profile2]
host = host2.example.com
port = 8002
`
	if err := os.WriteFile(configFile2, []byte(configContent2), 0600); err != nil {
		t.Fatal(err)
	}

	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl"}

	// Load first config
	loader1 := NewLoader().WithConfigFile(configFile1)
	cfg1, err := loader1.Load()
	if err != nil {
		t.Fatalf("Loading config1: %v", err)
	}

	// Load second config
	loader2 := NewLoader().WithConfigFile(configFile2)
	cfg2, err := loader2.Load()
	if err != nil {
		t.Fatalf("Loading config2: %v", err)
	}

	// Verify first config is unchanged
	if cfg1.Settings.WorkingDir != "/path/one" {
		t.Errorf("cfg1.Settings.WorkingDir = %q, want /path/one", cfg1.Settings.WorkingDir)
	}

	if cfg1.Settings.DefaultProfile != "profile1" {
		t.Errorf("cfg1.Settings.DefaultProfile = %q, want profile1", cfg1.Settings.DefaultProfile)
	}

	if cfg1.Settings.Git.Name != "User One" {
		t.Errorf("cfg1.Settings.Git.Name = %q, want User One", cfg1.Settings.Git.Name)
	}

	profile1, err := cfg1.GetProfile("profile1")
	if err != nil {
		t.Fatalf("cfg1.GetProfile(profile1): %v", err)
	}

	if profile1.Host != "host1.example.com" {
		t.Errorf("profile1.Host = %q, want host1.example.com", profile1.Host)
	}

	if profile1.Port != 8001 {
		t.Errorf("profile1.Port = %d, want 8001", profile1.Port)
	}

	// Verify second config has its own values
	if cfg2.Settings.WorkingDir != "/path/two" {
		t.Errorf("cfg2.Settings.WorkingDir = %q, want /path/two", cfg2.Settings.WorkingDir)
	}

	if cfg2.Settings.DefaultProfile != "profile2" {
		t.Errorf("cfg2.Settings.DefaultProfile = %q, want profile2", cfg2.Settings.DefaultProfile)
	}

	if cfg2.Settings.Git.Name != "User Two" {
		t.Errorf("cfg2.Settings.Git.Name = %q, want User Two", cfg2.Settings.Git.Name)
	}

	profile2, err := cfg2.GetProfile("profile2")
	if err != nil {
		t.Fatalf("cfg2.GetProfile(profile2): %v", err)
	}

	if profile2.Host != "host2.example.com" {
		t.Errorf("profile2.Host = %q, want host2.example.com", profile2.Host)
	}

	if profile2.Port != 8002 {
		t.Errorf("profile2.Port = %d, want 8002", profile2.Port)
	}

	// Verify cfg1 and cfg2 are completely independent
	if cfg1.Settings.WorkingDir == cfg2.Settings.WorkingDir {
		t.Error("cfg1 and cfg2 should have different working directories")
	}

	if cfg1.Settings.DefaultProfile == cfg2.Settings.DefaultProfile {
		t.Error("cfg1 and cfg2 should have different default profiles")
	}
}

// TestDetectConfigType verifies that detectConfigType correctly identifies file formats.
func TestDetectConfigType(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "YAML extension .yaml",
			filePath: "/path/to/config.yaml",
			want:     "yaml",
		},
		{
			name:     "YAML extension .yml",
			filePath: "/path/to/config.yml",
			want:     "yaml",
		},
		{
			name:     "TOML extension",
			filePath: "/path/to/config.toml",
			want:     "toml",
		},
		{
			name:     "JSON extension",
			filePath: "/path/to/config.json",
			want:     "json",
		},
		{
			name:     "INI extension",
			filePath: "/path/to/config.ini",
			want:     "ini",
		},
		{
			name:     "uppercase extension",
			filePath: "/path/to/config.YAML",
			want:     "yaml",
		},
		{
			name:     "mixed case extension",
			filePath: "/path/to/config.ToMl",
			want:     "toml",
		},
		{
			name:     "no extension defaults to ini",
			filePath: "/path/to/config",
			want:     "ini",
		},
		{
			name:     "unknown extension defaults to ini",
			filePath: "/path/to/config.txt",
			want:     "ini",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectConfigType(tt.filePath)
			if got != tt.want {
				t.Errorf("detectConfigType(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}

// TestLoaderLoadWithYAMLConfigFile verifies that YAML config files are loaded correctly.
func TestLoaderLoadWithYAMLConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `application:
  working_dir: /test/yaml/working
  default_profile: yamlprofile
  default_repository: yamlrepo

features:
  datasets_enabled: true

git:
  name: YAML User
  email: yaml@example.com
  user: yamlgit

profile default:
  host: yaml-default.example.com
  port: 9090
  use_tls: true
  verify: false
  username: yamldefaultuser
  password: yamldefaultpass

profile production:
  host: yaml-prod.example.com
  port: 8443
  use_tls: true
  verify: true
  username: yamlproduser

repository myrepo:
  url: https://github.com/example/yaml-repo.git
  reference: develop
`

	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write YAML config file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl"}

	loader := NewLoader().WithConfigFile(configFile)

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Verify application settings
	if cfg.Settings.WorkingDir != "/test/yaml/working" {
		t.Errorf("cfg.Settings.WorkingDir = %q, want %q", cfg.Settings.WorkingDir, "/test/yaml/working")
	}

	if cfg.Settings.DefaultProfile != "yamlprofile" {
		t.Errorf("cfg.Settings.DefaultProfile = %q, want %q", cfg.Settings.DefaultProfile, "yamlprofile")
	}

	// Verify git config
	if cfg.Settings.Git.Name != "YAML User" {
		t.Errorf("cfg.Settings.Git.Name = %q, want %q", cfg.Settings.Git.Name, "YAML User")
	}

	// Verify profiles
	defaultProfile, err := cfg.GetProfile("default")
	if err != nil {
		t.Fatalf("GetProfile(default) returned error: %v", err)
	}

	if defaultProfile.Host != "yaml-default.example.com" {
		t.Errorf("defaultProfile.Host = %q, want %q", defaultProfile.Host, "yaml-default.example.com")
	}

	if defaultProfile.Port != 9090 {
		t.Errorf("defaultProfile.Port = %d, want %d", defaultProfile.Port, 9090)
	}
}

// TestLoaderLoadWithTOMLConfigFile verifies that TOML config files are loaded correctly.
func TestLoaderLoadWithTOMLConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.toml")

	// Note: TOML requires profile sections to use spaces in the key name,
	// similar to INI format. The dot notation [profile.default] creates
	// a nested table structure which won't work with our profile loading logic.
	configContent := `[application]
working_dir = "/test/toml/working"
default_profile = "tomlprofile"
default_repository = "tomlrepo"

[features]
datasets_enabled = true

[git]
name = "TOML User"
email = "toml@example.com"
user = "tomlgit"

["profile default"]
host = "toml-default.example.com"
port = 7070
use_tls = true
verify = false
username = "tomldefaultuser"
password = "tomldefaultpass"

["profile production"]
host = "toml-prod.example.com"
port = 7443
use_tls = true
verify = true
username = "tomlproduser"

["repository myrepo"]
url = "https://github.com/example/toml-repo.git"
reference = "stable"
`

	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write TOML config file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl"}

	loader := NewLoader().WithConfigFile(configFile)

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Verify application settings
	if cfg.Settings.WorkingDir != "/test/toml/working" {
		t.Errorf("cfg.Settings.WorkingDir = %q, want %q", cfg.Settings.WorkingDir, "/test/toml/working")
	}

	if cfg.Settings.DefaultProfile != "tomlprofile" {
		t.Errorf("cfg.Settings.DefaultProfile = %q, want %q", cfg.Settings.DefaultProfile, "tomlprofile")
	}

	// Verify git config
	if cfg.Settings.Git.Name != "TOML User" {
		t.Errorf("cfg.Settings.Git.Name = %q, want %q", cfg.Settings.Git.Name, "TOML User")
	}

	// Verify profiles
	defaultProfile, err := cfg.GetProfile("default")
	if err != nil {
		t.Fatalf("GetProfile(default) returned error: %v", err)
	}

	if defaultProfile.Host != "toml-default.example.com" {
		t.Errorf("defaultProfile.Host = %q, want %q", defaultProfile.Host, "toml-default.example.com")
	}

	if defaultProfile.Port != 7070 {
		t.Errorf("defaultProfile.Port = %d, want %d", defaultProfile.Port, 7070)
	}
}

// TestLoaderLoadWithJSONConfigFile verifies that JSON config files are loaded correctly.
func TestLoaderLoadWithJSONConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	configContent := `{
  "application": {
    "working_dir": "/test/json/working",
    "default_profile": "jsonprofile",
    "default_repository": "jsonrepo"
  },
  "features": {
    "datasets_enabled": true
  },
  "git": {
    "name": "JSON User",
    "email": "json@example.com",
    "user": "jsongit"
  },
  "profile default": {
    "host": "json-default.example.com",
    "port": 6060,
    "use_tls": true,
    "verify": false,
    "username": "jsondefaultuser",
    "password": "jsondefaultpass"
  },
  "profile production": {
    "host": "json-prod.example.com",
    "port": 6443,
    "use_tls": true,
    "verify": true,
    "username": "jsonproduser"
  },
  "repository myrepo": {
    "url": "https://github.com/example/json-repo.git",
    "reference": "release"
  }
}`

	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write JSON config file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl"}

	loader := NewLoader().WithConfigFile(configFile)

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Verify application settings
	if cfg.Settings.WorkingDir != "/test/json/working" {
		t.Errorf("cfg.Settings.WorkingDir = %q, want %q", cfg.Settings.WorkingDir, "/test/json/working")
	}

	if cfg.Settings.DefaultProfile != "jsonprofile" {
		t.Errorf("cfg.Settings.DefaultProfile = %q, want %q", cfg.Settings.DefaultProfile, "jsonprofile")
	}

	// Verify git config
	if cfg.Settings.Git.Name != "JSON User" {
		t.Errorf("cfg.Settings.Git.Name = %q, want %q", cfg.Settings.Git.Name, "JSON User")
	}

	// Verify profiles
	defaultProfile, err := cfg.GetProfile("default")
	if err != nil {
		t.Fatalf("GetProfile(default) returned error: %v", err)
	}

	if defaultProfile.Host != "json-default.example.com" {
		t.Errorf("defaultProfile.Host = %q, want %q", defaultProfile.Host, "json-default.example.com")
	}

	if defaultProfile.Port != 6060 {
		t.Errorf("defaultProfile.Port = %d, want %d", defaultProfile.Port, 6060)
	}
}

// TestLoaderLoadMixedFormats verifies that different format files can be loaded independently.
func TestLoaderLoadMixedFormats(t *testing.T) {
	tmpDir := t.TempDir()

	// Create INI config
	iniFile := filepath.Join(tmpDir, "config.ini")
	iniContent := `[git]
name = INI User
`
	if err := os.WriteFile(iniFile, []byte(iniContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Create YAML config
	yamlFile := filepath.Join(tmpDir, "config.yaml")
	yamlContent := `git:
  name: YAML User
`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Create TOML config
	tomlFile := filepath.Join(tmpDir, "config.toml")
	tomlContent := `[git]
name = "TOML User"
`
	if err := os.WriteFile(tomlFile, []byte(tomlContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Create JSON config
	jsonFile := filepath.Join(tmpDir, "config.json")
	jsonContent := `{"git": {"name": "JSON User"}}`
	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0600); err != nil {
		t.Fatal(err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ipctl"}

	// Test each format
	tests := []struct {
		name     string
		file     string
		wantName string
	}{
		{"INI format", iniFile, "INI User"},
		{"YAML format", yamlFile, "YAML User"},
		{"TOML format", tomlFile, "TOML User"},
		{"JSON format", jsonFile, "JSON User"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader().WithConfigFile(tt.file)
			cfg, err := loader.Load()
			if err != nil {
				t.Fatalf("Load() returned error: %v", err)
			}

			if cfg.Settings.Git.Name != tt.wantName {
				t.Errorf("cfg.Settings.Git.Name = %q, want %q", cfg.Settings.Git.Name, tt.wantName)
			}
		})
	}
}
