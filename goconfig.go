// Package goconfig provides a lightweight configuration loader for Go applications.
// It reads configuration files from a directory, supports environment variable
// substitution using the ${VAR_NAME} syntax, and can load .env files.
package goconfig

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	excludeExtensions = []string{"go"}
	regexEnv          = regexp.MustCompile(`\${(\w+)}`)
	regexEnvFromFile  = regexp.MustCompile(`^\s*([\w.-]+)\s*=\s*(.*)?\s*$`)
)

// Exported errors that can be used by callers to identify specific failure modes.
var (
	// ErrVariableNotFound is returned when an environment variable referenced
	// with ${VAR} syntax is not defined in the environment.
	ErrVariableNotFound = fmt.Errorf("environment variable not found")

	// ErrInvalidEnvFormat is returned when a line in an .env file does not
	// match the expected KEY=value format.
	ErrInvalidEnvFormat = fmt.Errorf("invalid environment variable format")
)

// Unmarshaller is a function type that unmarshals raw configuration bytes
// into the provided structure. The default unmarshaller parses YAML.
type Unmarshaller func(interface{}, []byte) error

// Config holds the configuration state for the loader, including the directory
// to scan and the unmarshaller function to use.
type Config struct {
	configDir      string
	unmarshallFunc Unmarshaller
}

// Option is a functional option that configures a Config instance.
type Option func(*Config)

// WithConfigDir sets the directory where configuration files will be searched.
// The default directory is "config".
func WithConfigDir(dir string) Option {
	return func(c *Config) {
		c.configDir = dir
	}
}

// WithUnmarshaller sets a custom unmarshaller function for parsing configuration files.
// This can be used to support formats other than YAML, such as JSON or TOML.
func WithUnmarshaller(u Unmarshaller) Option {
	return func(c *Config) {
		c.unmarshallFunc = u
	}
}

// New creates a new Config with the provided options.
// By default, it scans the "config" directory and uses a YAML unmarshaller.
func New(opts ...Option) *Config {
	c := &Config{
		configDir:      "config",
		unmarshallFunc: unmarshallYAML,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// LoadEnv reads environment variables from the given .env files and loads them
// into the current process environment. If no files are provided, it defaults to ".env".
func (c *Config) LoadEnv(envFiles ...string) error {
	if len(envFiles) == 0 {
		envFiles = []string{".env"}
	}
	for _, envFile := range envFiles {
		if err := c.parseEnvFile(envFile); err != nil {
			return err
		}
	}
	return nil
}

// Parse searches for a configuration file matching fileName (case-insensitive, without extension)
// in the configured directory. It replaces environment variables using ${VAR} syntax and unmarshals
// the content into the provided structure.
func (c *Config) Parse(fileName string, structure interface{}) error {
	content, err := c.read(fileName)
	if err != nil {
		return err
	}
	contentStr, err := replaceEnvVariables(string(content))
	if err != nil {
		return fmt.Errorf("failed to replace env variables in %s: %w", fileName, err)
	}
	if err := c.unmarshallFunc(structure, []byte(contentStr)); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", fileName, err)
	}
	return nil
}

// Read scans the config directory for a file whose base name matches fileName
// (case-insensitive) and returns its raw contents. Files without extensions or
// with the "go" extension are ignored.
func (c *Config) read(fileName string) ([]byte, error) {
	files, err := os.ReadDir(c.configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory '%s': %w", c.configDir, err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name, extension, found := strings.Cut(file.Name(), ".")
		if !found {
			continue
		}
		if slices.Contains(excludeExtensions, extension) {
			continue
		}
		if strings.EqualFold(name, fileName) {
			fullPath := filepath.Join(c.configDir, file.Name())
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file '%s': %w", fullPath, err)
			}
			return content, nil
		}
	}
	return nil, fmt.Errorf("configuration file '%s' not found in '%s'", fileName, c.configDir)
}

// replaceEnvVariables replaces all occurrences of ${VAR} in the given content
// with the corresponding environment variable value. If a variable is not found,
// it returns ErrVariableNotFound.
func replaceEnvVariables(content string) (string, error) {
	var err error
	result := regexEnv.ReplaceAllStringFunc(content, func(match string) string {
		envVar := regexEnv.FindStringSubmatch(match)[1]
		env, found := os.LookupEnv(envVar)
		if !found {
			err = fmt.Errorf("%w: %s", ErrVariableNotFound, envVar)
			return match
		}
		return env
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

// unmarshallYAML is the default unmarshaller that parses YAML content into
// the provided Go structure using gopkg.in/yaml.v3.
func unmarshallYAML(structure interface{}, content []byte) error {
	return yaml.Unmarshal(content, structure)
}

// parseEnvFile opens the given .env file, reads it line by line, and sets each
// valid KEY=value pair as an environment variable using os.Setenv.
func (c *Config) parseEnvFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open env file '%s': %w", filePath, err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if isCommentOrEmpty(line) {
			continue
		}
		if err := setEnvVarFromLine(line); err != nil {
			return fmt.Errorf("error in '%s': %w", filePath, err)
		}
	}
	return scanner.Err()
}

// isCommentOrEmpty reports whether the given line is empty or starts with a comment.
func isCommentOrEmpty(line string) bool {
	trimmed := strings.TrimSpace(line)
	return len(trimmed) == 0 || strings.HasPrefix(trimmed, "#")
}

// setEnvVarFromLine parses a single KEY=value line and sets the environment
// variable using os.Setenv. It supports optional single or double quotes around
// the value, which are stripped before setting.
func setEnvVarFromLine(line string) error {
	matches := regexEnvFromFile.FindStringSubmatch(line)
	if len(matches) != 3 {
		return fmt.Errorf("%w: %s", ErrInvalidEnvFormat, line)
	}
	key := matches[1]
	value := matches[2]
	if len(value) >= 2 {
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}
	}
	return os.Setenv(key, value)
}
