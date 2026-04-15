# dexconfig

`dexconfig` is a small Go package for loading configuration from environment variables into a struct.

It uses reflection and `env` tags to populate fields, supports nested structs, pointer-to-struct fields, default values, and `time.Duration`.

## Install

```bash
go get github.com/barluscuda/dexconfig/dexconfig
```

## Usage

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/barluscuda/dexconfig/dexconfig"
)

type DatabaseConfig struct {
	Host string `env:"DB_HOST;localhost"`
	Port int    `env:"DB_PORT;5432"`
}

type AppConfig struct {
	Name    string         `env:"APP_NAME;dexconfig"`
	Debug   bool           `env:"APP_DEBUG;false"`
	Timeout time.Duration  `env:"APP_TIMEOUT;5s"`
	DB      DatabaseConfig
}

func main() {
	var cfg AppConfig

	if err := dexconfig.LoadConfig(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
}
```

Example environment:

```bash
export APP_NAME=my-service
export APP_DEBUG=true
export APP_TIMEOUT=10s
export DB_HOST=db.internal
export DB_PORT=5432
```

## Tag Format

Use the `env` struct tag:

```go
Field string `env:"ENV_KEY;default-value"`
```

- `ENV_KEY` is the environment variable name.
- `default-value` is optional.

Examples:

```go
Host string `env:"HOST"`
Port int    `env:"PORT;8080"`
Debug bool  `env:"DEBUG;false"`
```

## Supported Field Types

- `string`
- Signed integer types: `int`, `int8`, `int16`, `int32`, `int64`
- `bool`
- `float32`, `float64`
- `time.Duration`

## Nested Structs

Nested structs are loaded recursively.

```go
type TLSConfig struct {
	Enabled bool `env:"TLS_ENABLED;false"`
}

type Config struct {
	TLS TLSConfig
}
```

Pointer-to-struct fields are also supported and are initialized automatically before loading:

```go
type Config struct {
	TLS *TLSConfig
}
```

## Behavior

- `LoadConfig` requires a non-nil pointer to a struct.
- Fields without an `env` tag are ignored.
- Nested structs are traversed recursively.
- If an environment variable is missing or resolves to an empty string, the default from the tag is used.
- If no value is available and the target type cannot parse an empty string, `LoadConfig` returns an error.
- Errors include the field name that failed to load.

## Limitations

Current implementation does not support:

- Unsigned integers
- Slices and arrays
- Maps
- Custom parsing hooks

Unsupported tagged fields return an error such as `unsupport type: ...`.

## API

```go
func LoadConfig(c interface{}) error
```

## Example Error Cases

`LoadConfig` returns an error when:

- The input is not a pointer
- The input pointer is nil
- The pointer does not reference a struct
- A tagged field contains a value that cannot be parsed into its target type

## Developer

Developed by [barluscuda](https://github.com/barluscuda).

## License

This project is licensed under the MIT License. See [LICENSE](/home/mrbarlus/projects/dexconfig/LICENSE) for details.
