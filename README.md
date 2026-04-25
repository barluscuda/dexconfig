# dexconfig

`dexconfig` is a lightweight, dependency-free Go package for loading configuration
from environment variables into structs.

It uses reflection and struct tags to populate fields, with support for:

* Nested structs
* Pointer-to-struct fields (auto-initialized)
* Default values
* Required fields
* Slices and maps
* `time.Duration`, `time.Time`
* Any type implementing `encoding.TextUnmarshaler`

---

## Install

```bash
go get github.com/barluscuda/dexconfig
```

---

## Quick Start

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/barluscuda/dexconfig"
)

type DatabaseConfig struct {
	Host string `env:"DB_HOST=localhost"`
	Port int    `env:"DB_PORT=5432"`
}

type AppConfig struct {
	Name    string        `env:"APP_NAME=dexconfig"`
	Debug   bool          `env:"APP_DEBUG=false"`
	Timeout time.Duration `env:"APP_TIMEOUT=5s"`
	Secret  string        `env:"APP_SECRET" envrequired:"true"`
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

---

## Tag Format

```go
Field T `env:"ENV_KEY=default"`                    // optional with default
Field T `env:"ENV_KEY"`                            // optional, no default
Field T `env:"ENV_KEY" envrequired:"true"`         // required
Field T `env:"ENV_KEY=default" envrequired:"true"` // required with fallback
Field T `env:"-"`                                  // ignore field
```

### Rules

* `ENV_KEY` → environment variable name
* `=default` → used if variable is unset
* `envrequired:"true"` → fails if missing or empty
* `-` → skips the field

---

## Options

```go
dexconfig.LoadConfig(&cfg,
	dexconfig.WithPrefix("MYAPP"),      // MYAPP_<KEY>
	dexconfig.WithTagName("env"),       // custom tag name
	dexconfig.WithSeparator(","),       // slice/map separator
	dexconfig.WithLookup(os.LookupEnv), // custom env source (testing)
)
```

---

## Supported Types

```go
type Example struct {
	Name string `env:"NAME=dexconfig"`

	// integers
	Port  int   `env:"PORT=8080"`
	Small int8  `env:"SMALL=-1"`

	// unsigned
	Retries uint32 `env:"RETRIES=3"`

	// bool
	Debug bool `env:"DEBUG=false"`

	// float
	Ratio float64 `env:"RATIO=0.5"`

	// time
	Timeout time.Duration `env:"TIMEOUT=5s"`
	StartAt time.Time     `env:"START_AT=2026-04-15T00:00:00Z"`

	// slices
	Tags  []string `env:"TAGS"`  // TAGS=a,b,c
	Ports []int    `env:"PORTS"` // PORTS=80,443

	// maps
	Labels map[string]string `env:"LABELS"` // LABELS=env:prod,tier:web

	// custom types
	Addr netip.Addr `env:"ADDR=0.0.0.0"`

	// nested
	DB DBConfig

	// pointer (auto-init)
	TLS *TLSConfig
}
```

---

## Error Handling

```go
var fe *dexconfig.FieldError
if errors.As(err, &fe) {
	log.Printf("env %s failed: %v", fe.Key, fe.Err)
}
```

---

## Notes

* Empty string is treated as **missing** when `envrequired:"true"`
* Pointer fields are automatically initialized
* Supports nested structs with prefix
* Slice/map separator is configurable (default: `,`)

---

## License

MIT — see [LICENSE](LICENSE)
