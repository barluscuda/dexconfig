# dexconfig

`dexconfig` is a small, dependency-free Go package for loading configuration
from environment variables into a struct.

It uses reflection and `env` tags to populate fields, supports nested
structs, pointer-to-struct fields, default values, required fields, slices,
maps, `time.Duration`, `time.Time`, and any type implementing
`encoding.TextUnmarshaler`.

## Install

```bash
go get github.com/barluscuda/dexconfig
```

## Usage

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/barluscuda/dexconfig"
)

type DatabaseConfig struct {
	Host string `env:"DB_HOST:localhost"`
	Port int    `env:"DB_PORT:5432"`
}

type AppConfig struct {
	Name    string        `env:"APP_NAME:dexconfig"`
	Debug   bool          `env:"APP_DEBUG:false"`
	Timeout time.Duration `env:"APP_TIMEOUT:5s"`
	Secret  string        `env:"APP_SECRET:;required"`
	Tags    []string      `env:"APP_TAGS"`
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

## Tag Format

```go
Field T `env:"ENV_KEY:default-value"`
Field T `env:"ENV_KEY:;required"`
Field T `env:"-"` // skip
```

- `ENV_KEY` is the environment variable name.
- The portion after `:` is the default value (optional).
- Append `;required` to fail loading when the variable is unset/empty.
- Use `-` to ignore a field.

## Options

```go
dexconfig.LoadConfig(&cfg,
    dexconfig.WithPrefix("MYAPP"),     // looks up MYAPP_<KEY>
    dexconfig.WithTagName("env"),      // override struct tag name
    dexconfig.WithSeparator(","),      // slice/map separator
    dexconfig.WithLookup(os.LookupEnv) // override env source (useful in tests)
)
```

## Supported Field Types

```go
type Example struct {
    // string
    Name string `env:"NAME:dexconfig"`

    // signed integers: int, int8, int16, int32, int64
    Port  int   `env:"PORT:8080"`
    Small int8  `env:"SMALL:-1"`

    // unsigned integers: uint, uint8, uint16, uint32, uint64
    Retries uint32 `env:"RETRIES:3"`

    // bool
    Debug bool `env:"DEBUG:false"`

    // float32, float64
    Ratio float64 `env:"RATIO:0.5"`

    // time.Duration
    Timeout time.Duration `env:"TIMEOUT:5s"`

    // time.Time (RFC3339)
    StartAt time.Time `env:"START_AT:2026-04-15T00:00:00Z"`

    // slices of supported scalars (comma-separated by default)
    Tags  []string `env:"TAGS"`          // TAGS=a,b,c
    Ports []int    `env:"PORTS"`         // PORTS=80,443

    // maps of supported scalars, formatted as key:val,key:val
    Labels map[string]string `env:"LABELS"` // LABELS=env:prod,tier:web

    // any type implementing encoding.TextUnmarshaler
    Addr netip.Addr `env:"ADDR:0.0.0.0"`

    // nested struct
    DB DBConfig

    // pointer-to-struct (initialized automatically)
    TLS *TLSConfig
}
```

## Errors

`LoadConfig` returns wrapped `*FieldError` values with the failing field name
and environment key. Use `errors.As` to inspect them:

```go
var fe *dexconfig.FieldError
if errors.As(err, &fe) {
    log.Printf("env %s failed: %v", fe.Key, fe.Err)
}
```

## License

MIT — see [LICENSE](LICENSE).
