# goconfig

[![Go Reference](https://pkg.go.dev/badge/github.com/salomondevsystems/goconfig.svg)](https://pkg.go.dev/github.com/salomondevsystems/goconfig)
[![Go Report Card](https://goreportcard.com/badge/github.com/salomondevsystems/goconfig)](https://goreportcard.com/report/github.com/salomondevsystems/goconfig)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`goconfig` is a lightweight Go library for loading application configuration from files (YAML by default) with built-in support for environment variable substitution using the `${VAR_NAME}` syntax. It also includes utilities for loading `.env` files.

## Features

- **Declarative configuration loading**: Reads configuration files from a directory and unmarshals them into a Go struct.
- **Environment variable substitution**: Automatically replaces `${VAR}` placeholders with the corresponding environment variable values.
- **`.env` file support**: Loads environment variables from `.env` files before parsing the configuration.
- **Customizable**: Allows changing the configuration directory and the unmarshaller (e.g., to support JSON or TOML).
- **Lightweight dependencies**: Only requires `gopkg.in/yaml.v3`.

## Installation

```bash
go get github.com/salomondevsystems/goconfig
```

## Quick Start

### 1. Define your configuration struct

```go
type Config struct {
    App struct {
        Name string `yaml:"name"`
        Port int    `yaml:"port"`
    } `yaml:"app"`
}
```

### 2. Create your configuration file (`config/config.yaml`)

```yaml
app:
  name: ${APP_NAME}
  port: ${APP_PORT}
```

### 3. Load the configuration in your application

```go
package main

import (
    "fmt"
    "log"
    "github.com/salomondevsystems/goconfig"
)

type Config struct {
    App struct {
        Name string `yaml:"name"`
        Port int    `yaml:"port"`
    } `yaml:"app"`
}

func main() {
    // Load variables from .env (optional)
    cfg := goconfig.New()
    if err := cfg.LoadEnv(".env"); err != nil {
        log.Fatalf("failed to load env: %v", err)
    }

    // Parse the configuration file
    var appConfig Config
    if err := cfg.Parse("config", &appConfig); err != nil {
        log.Fatalf("failed to parse config: %v", err)
    }

    fmt.Printf("App: %s on port %d\n", appConfig.App.Name, appConfig.App.Port)
}
```

## Options

### Change the configuration directory

By default, `goconfig` looks for files in the `config` directory. You can change it:

```go
cfg := goconfig.New(goconfig.WithConfigDir("/path/to/configs"))
```

### Use a custom unmarshaller

If you need to support other formats (JSON, TOML, etc.), you can provide your own unmarshalling function:

```go
import "encoding/json"

cfg := goconfig.New(goconfig.WithUnmarshaller(func(v interface{}, data []byte) error {
    return json.Unmarshal(data, v)
}))
```

### Load multiple `.env` files

```go
if err := cfg.LoadEnv(".env", ".env.local", ".env.production"); err != nil {
    log.Fatal(err)
}
```

## API

### `func New(opts ...Option) *Config`
Creates a new configuration instance with the provided options.

### `func (c *Config) LoadEnv(envFiles ...string) error`
Loads environment variables from one or more `.env` files. If no files are provided, it defaults to `.env`.

### `func (c *Config) Parse(fileName string, structure interface{}) error`
Searches for a file by name (without extension) in the configuration directory, replaces environment variables `${VAR}`, and unmarshals the content into the provided structure.

### Exported Errors
- `ErrVariableNotFound`: Returned when an environment variable referenced with `${VAR}` is not defined.
- `ErrInvalidEnvFormat`: Returned when a line in an `.env` file has an invalid format.

## Full Example

See the [`examples/basic/`](./examples/basic/) directory for a working, runnable example.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for the full history of changes.
