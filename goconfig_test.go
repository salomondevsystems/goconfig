package goconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// --- Tests for New() ---

func TestNew_DefaultOptions(t *testing.T) {
	cfg := New()
	if cfg.configDir != "config" {
		t.Errorf("expected configDir 'config', got '%s'", cfg.configDir)
	}
	if cfg.unmarshallFunc == nil {
		t.Error("expected unmarshallFunc to be set")
	}
}

func TestNew_WithConfigDir(t *testing.T) {
	cfg := New(WithConfigDir("custom/config"))
	if cfg.configDir != "custom/config" {
		t.Errorf("expected configDir 'custom/config', got '%s'", cfg.configDir)
	}
}

func TestNew_WithUnmarshaller(t *testing.T) {
	custom := func(interface{}, []byte) error { return nil }
	cfg := New(WithUnmarshaller(custom))
	if cfg.unmarshallFunc == nil {
		t.Error("expected unmarshallFunc to be set")
	}
}

// --- Tests for LoadEnv() ---

func TestLoadEnv_DefaultFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	content := "TEST_VAR_DEFAULT=hello\n"
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := New()
	// Change to the temporary directory so that .env is found
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()

	if err := cfg.LoadEnv(); err != nil {
		t.Fatalf("LoadEnv() error = %v", err)
	}

	if os.Getenv("TEST_VAR_DEFAULT") != "hello" {
		t.Errorf("expected TEST_VAR_DEFAULT='hello', got '%s'", os.Getenv("TEST_VAR_DEFAULT"))
	}
}

func TestLoadEnv_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	env1 := filepath.Join(tmpDir, "a.env")
	env2 := filepath.Join(tmpDir, "b.env")

	if err := os.WriteFile(env1, []byte("VAR_A=first\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(env2, []byte("VAR_B=second\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := New()
	if err := cfg.LoadEnv(env1, env2); err != nil {
		t.Fatalf("LoadEnv() error = %v", err)
	}

	if os.Getenv("VAR_A") != "first" {
		t.Errorf("expected VAR_A='first', got '%s'", os.Getenv("VAR_A"))
	}
	if os.Getenv("VAR_B") != "second" {
		t.Errorf("expected VAR_B='second', got '%s'", os.Getenv("VAR_B"))
	}
}

func TestLoadEnv_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "bad.env")
	if err := os.WriteFile(envFile, []byte("THIS_IS_NOT_VALID\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := New()
	err := cfg.LoadEnv(envFile)
	if err == nil {
		t.Fatal("expected error for invalid env format")
	}
}

func TestLoadEnv_FileNotFound(t *testing.T) {
	cfg := New()
	err := cfg.LoadEnv("/nonexistent/path/.env")
	if err == nil {
		t.Fatal("expected error when env file not found")
	}
}

// --- Tests for Parse() ---

func TestParse_Success(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "app.yaml")
	content := `
name: TestApp
port: 9090
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var result struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	}

	cfg := New(WithConfigDir(tmpDir))
	if err := cfg.Parse("app", &result); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.Name != "TestApp" {
		t.Errorf("expected Name='TestApp', got '%s'", result.Name)
	}
	if result.Port != 9090 {
		t.Errorf("expected Port=9090, got %d", result.Port)
	}
}

func TestParse_EnvSubstitution(t *testing.T) {
	t.Setenv("APP_NAME", "SubstitutedApp")
	t.Setenv("APP_PORT", "7777")

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "app.yaml")
	content := `
name: ${APP_NAME}
port: ${APP_PORT}
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var result struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	}

	cfg := New(WithConfigDir(tmpDir))
	if err := cfg.Parse("app", &result); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.Name != "SubstitutedApp" {
		t.Errorf("expected Name='SubstitutedApp', got '%s'", result.Name)
	}
	if result.Port != 7777 {
		t.Errorf("expected Port=7777, got %d", result.Port)
	}
}

func TestParse_MissingEnvVar(t *testing.T) {
	if err := os.Unsetenv("MISSING_VAR_XYZ"); err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "app.yaml")
	content := `name: ${MISSING_VAR_XYZ}`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var result struct {
		Name string `yaml:"name"`
	}

	cfg := New(WithConfigDir(tmpDir))
	err := cfg.Parse("app", &result)
	if err == nil {
		t.Fatal("expected error for missing env variable")
	}
}

func TestParse_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	var result struct{}
	cfg := New(WithConfigDir(tmpDir))

	err := cfg.Parse("nonexistent", &result)
	if err == nil {
		t.Fatal("expected error when file not found")
	}
}

func TestParse_CustomUnmarshaller(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "app.json")
	content := `{"name": "JsonApp", "port": 3000}`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var result struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	}

	customUnmarshaller := func(v interface{}, data []byte) error {
		return json.Unmarshal(data, v)
	}

	cfg := New(WithConfigDir(tmpDir), WithUnmarshaller(customUnmarshaller))
	if err := cfg.Parse("app", &result); err != nil {
		t.Fatalf("Parse() with custom unmarshaller error = %v", err)
	}

	if result.Name != "JsonApp" {
		t.Errorf("expected Name='JsonApp', got '%s'", result.Name)
	}
	if result.Port != 3000 {
		t.Errorf("expected Port=3000, got %d", result.Port)
	}
}

func TestParse_IgnoresGoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a .go file with the same base name
	if err := os.WriteFile(filepath.Join(tmpDir, "app.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	// Do not create app.yaml, so it should fail

	var result struct{}
	cfg := New(WithConfigDir(tmpDir))

	err := cfg.Parse("app", &result)
	if err == nil {
		t.Fatal("expected error, .go files should be ignored")
	}
}

func TestParse_CaseInsensitiveMatch(t *testing.T) {
	tmpDir := t.TempDir()
	// File with uppercase extension
	if err := os.WriteFile(filepath.Join(tmpDir, "APP.YAML"), []byte("name: caseapp\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var result struct {
		Name string `yaml:"name"`
	}

	cfg := New(WithConfigDir(tmpDir))
	if err := cfg.Parse("app", &result); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.Name != "caseapp" {
		t.Errorf("expected Name='caseapp', got '%s'", result.Name)
	}
}

// --- Tests for replaceEnvVariables() ---

func TestReplaceEnvVariables(t *testing.T) {
	t.Setenv("FOO", "bar")
	t.Setenv("NUM", "42")

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "single substitution",
			input:    "value: ${FOO}",
			expected: "value: bar",
			wantErr:  false,
		},
		{
			name:     "multiple substitutions",
			input:    "a: ${FOO}, b: ${NUM}",
			expected: "a: bar, b: 42",
			wantErr:  false,
		},
		{
			name:     "no substitution",
			input:    "static value",
			expected: "static value",
			wantErr:  false,
		},
		{
			name:    "missing variable",
			input:   "value: ${UNKNOWN_VAR_123}",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := replaceEnvVariables(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("replaceEnvVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("replaceEnvVariables() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// --- Tests for isCommentOrEmpty() ---

func TestIsCommentOrEmpty(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"   ", true},
		{"# comment", true},
		{"  # comment with spaces", true},
		{"KEY=value", false},
		{"  KEY=value  ", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			got := isCommentOrEmpty(tt.input)
			if got != tt.expected {
				t.Errorf("isCommentOrEmpty(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// --- Tests for setEnvVarFromLine() ---

func TestSetEnvVarFromLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "simple key=value",
			line:      "KEY=value",
			wantKey:   "KEY",
			wantValue: "value",
			wantErr:   false,
		},
		{
			name:      "with spaces",
			line:      "  KEY = value",
			wantKey:   "KEY",
			wantValue: "value",
			wantErr:   false,
		},
		{
			name:      "double quoted",
			line:      `KEY="quoted value"`,
			wantKey:   "KEY",
			wantValue: "quoted value",
			wantErr:   false,
		},
		{
			name:      "single quoted",
			line:      "KEY='single quoted'",
			wantKey:   "KEY",
			wantValue: "single quoted",
			wantErr:   false,
		},
		{
			name:      "empty value",
			line:      "KEY=",
			wantKey:   "KEY",
			wantValue: "",
			wantErr:   false,
		},
		{
			name:    "invalid format",
			line:    "INVALID_LINE",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear the variable if it exists
			if err := os.Unsetenv(tt.wantKey); err != nil {
				t.Fatal(err)
			}

			err := setEnvVarFromLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("setEnvVarFromLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got := os.Getenv(tt.wantKey)
				if got != tt.wantValue {
					t.Errorf("setEnvVarFromLine() got env %s=%q, want %q", tt.wantKey, got, tt.wantValue)
				}
			}
		})
	}
}

// --- Integration-like test ---

func TestFullFlow(t *testing.T) {
	// 1. Setup temp dir and files
	tmpDir := t.TempDir()

	envContent := `
DB_HOST=localhost
DB_PORT=5432
`
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	configContent := `
database:
  host: ${DB_HOST}
  port: ${DB_PORT}
app:
  name: FullFlowApp
`
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "app.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Load env and parse config
	cfg := New(WithConfigDir(configDir))
	if err := cfg.LoadEnv(envFile); err != nil {
		t.Fatalf("LoadEnv() error = %v", err)
	}

	var result struct {
		Database struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"database"`
		App struct {
			Name string `yaml:"name"`
		} `yaml:"app"`
	}

	if err := cfg.Parse("app", &result); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.Database.Host != "localhost" {
		t.Errorf("expected DB_HOST='localhost', got '%s'", result.Database.Host)
	}
	if result.Database.Port != 5432 {
		t.Errorf("expected DB_PORT=5432, got %d", result.Database.Port)
	}
	if result.App.Name != "FullFlowApp" {
		t.Errorf("expected App.Name='FullFlowApp', got '%s'", result.App.Name)
	}
}

// --- Tests for uncovered lines ---

func TestParse_UnmarshallError(t *testing.T) {
	tmpDir := t.TempDir()
	// Invalid YAML: list where a map is expected
	if err := os.WriteFile(filepath.Join(tmpDir, "bad.yaml"), []byte("- item1\n- item2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var result struct {
		Name string `yaml:"name"`
	}

	cfg := New(WithConfigDir(tmpDir))
	err := cfg.Parse("bad", &result)
	if err == nil {
		t.Fatal("expected error when unmarshalling invalid YAML")
	}
}

func TestRead_ConfigDirNotFound(t *testing.T) {
	cfg := New(WithConfigDir("/nonexistent/directory/12345"))
	var result struct{}

	err := cfg.Parse("anything", &result)
	if err == nil {
		t.Fatal("expected error when config directory does not exist")
	}
}

func TestRead_SkipsDirAndFilesWithoutExtension(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a subdirectory (must be skipped by IsDir)
	if err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	// Create a file without extension (must be skipped by !found)
	if err := os.WriteFile(filepath.Join(tmpDir, "noextension"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	// Do not create any matching config file to force the full loop

	var result struct{}
	cfg := New(WithConfigDir(tmpDir))
	err := cfg.Parse("app", &result)
	if err == nil {
		t.Fatal("expected error when config file not found")
	}
}

func TestRead_SkipsDirAndFilesWithoutExtension_WithMatch(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a subdirectory
	if err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	// Create a file without extension
	if err := os.WriteFile(filepath.Join(tmpDir, "noextension"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	// Create the valid config file
	if err := os.WriteFile(filepath.Join(tmpDir, "app.yaml"), []byte("name: valid\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var result struct {
		Name string `yaml:"name"`
	}

	cfg := New(WithConfigDir(tmpDir))
	if err := cfg.Parse("app", &result); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.Name != "valid" {
		t.Errorf("expected Name='valid', got '%s'", result.Name)
	}
}

func TestRead_FileReadError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "secret.yaml")
	if err := os.WriteFile(configFile, []byte("name: secret\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Remove read permissions
	if err := os.Chmod(configFile, 0000); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chmod(configFile, 0644); err != nil {
			t.Fatal(err)
		}
	}() // Restore so t.TempDir() can clean up

	var result struct {
		Name string `yaml:"name"`
	}

	cfg := New(WithConfigDir(tmpDir))
	err := cfg.Parse("secret", &result)
	if err == nil {
		t.Fatal("expected error when file cannot be read")
	}
}
