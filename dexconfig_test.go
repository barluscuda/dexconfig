package dexconfig

import (
	"errors"
	"net/netip"
	"testing"
	"time"
)

type DBConfig struct {
	Host string `env:"DB_HOST:localhost"`
	Port int    `env:"DB_PORT:5432"`
}

type TLSConfig struct {
	Enabled bool `env:"TLS_ENABLED:false"`
}

type FullConfig struct {
	Name    string            `env:"APP_NAME:dexconfig"`
	Debug   bool              `env:"APP_DEBUG:false"`
	Timeout time.Duration     `env:"APP_TIMEOUT:5s"`
	Secret  string            `env:"APP_SECRET:;required"`
	Tags    []string          `env:"APP_TAGS"`
	Ports   []int             `env:"APP_PORTS"`
	Labels  map[string]string `env:"APP_LABELS"`
	Addr    netip.Addr        `env:"APP_ADDR"`
	When    time.Time         `env:"APP_WHEN"`
	Ratio   float64           `env:"APP_RATIO:0.5"`
	Retries uint32            `env:"APP_RETRIES:3"`
	Skipped string            `env:"-"`
	DB      DBConfig
	TLS     *TLSConfig
}

func staticEnv(m map[string]string) LookupFunc {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	var c FullConfig
	err := LoadConfig(&c, WithLookup(staticEnv(map[string]string{
		"APP_SECRET": "shhh",
	})))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Name != "dexconfig" || c.Timeout != 5*time.Second || c.DB.Port != 5432 || c.Ratio != 0.5 || c.Retries != 3 {
		t.Fatalf("defaults not applied: %+v", c)
	}
	if c.TLS == nil || c.TLS.Enabled {
		t.Fatalf("nested pointer struct not initialized: %+v", c.TLS)
	}
}

func TestLoadConfig_Overrides(t *testing.T) {
	var c FullConfig
	err := LoadConfig(&c, WithLookup(staticEnv(map[string]string{
		"APP_NAME":    "svc",
		"APP_DEBUG":   "true",
		"APP_TIMEOUT": "2m",
		"APP_SECRET":  "x",
		"APP_TAGS":    "a, b ,c",
		"APP_PORTS":   "80,443",
		"APP_LABELS":  "env:prod, tier:web",
		"APP_ADDR":    "10.0.0.1",
		"APP_WHEN":    "2026-04-15T00:00:00Z",
		"DB_HOST":     "db",
		"DB_PORT":     "6543",
		"TLS_ENABLED": "true",
	})))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Name != "svc" || !c.Debug || c.Timeout != 2*time.Minute {
		t.Fatalf("scalars not parsed: %+v", c)
	}
	if len(c.Tags) != 3 || c.Tags[1] != "b" {
		t.Fatalf("tags not parsed: %+v", c.Tags)
	}
	if len(c.Ports) != 2 || c.Ports[1] != 443 {
		t.Fatalf("ports not parsed: %+v", c.Ports)
	}
	if c.Labels["env"] != "prod" || c.Labels["tier"] != "web" {
		t.Fatalf("labels not parsed: %+v", c.Labels)
	}
	if c.Addr.String() != "10.0.0.1" {
		t.Fatalf("addr not parsed: %+v", c.Addr)
	}
	if c.When.Year() != 2026 {
		t.Fatalf("time not parsed: %+v", c.When)
	}
	if c.DB.Host != "db" || c.DB.Port != 6543 || !c.TLS.Enabled {
		t.Fatalf("nested not parsed: %+v %+v", c.DB, c.TLS)
	}
}

func TestLoadConfig_Required(t *testing.T) {
	var c FullConfig
	err := LoadConfig(&c, WithLookup(staticEnv(map[string]string{})))
	if err == nil {
		t.Fatal("expected required error")
	}
	var fe *FieldError
	if !errors.As(err, &fe) || fe.Key != "APP_SECRET" {
		t.Fatalf("expected FieldError for APP_SECRET, got: %v", err)
	}
}

func TestLoadConfig_Prefix(t *testing.T) {
	type C struct {
		Host string `env:"HOST:default"`
	}
	var c C
	err := LoadConfig(&c, WithPrefix("MYAPP"), WithLookup(staticEnv(map[string]string{
		"MYAPP_HOST": "h",
	})))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if c.Host != "h" {
		t.Fatalf("prefix lookup failed: %+v", c)
	}
}

func TestLoadConfig_InvalidInput(t *testing.T) {
	if err := LoadConfig(nil); err == nil {
		t.Fatal("expected error for nil")
	}
	var s struct{}
	if err := LoadConfig(s); err == nil {
		t.Fatal("expected error for non-pointer")
	}
	var p *struct{}
	if err := LoadConfig(p); err == nil {
		t.Fatal("expected error for nil pointer")
	}
	x := 5
	if err := LoadConfig(&x); err == nil {
		t.Fatal("expected error for non-struct pointer")
	}
}

func TestLoadConfig_ParseError(t *testing.T) {
	type C struct {
		Port int `env:"PORT"`
	}
	var c C
	err := LoadConfig(&c, WithLookup(staticEnv(map[string]string{"PORT": "not-a-number"})))
	if err == nil {
		t.Fatal("expected parse error")
	}
	var fe *FieldError
	if !errors.As(err, &fe) || fe.Key != "PORT" {
		t.Fatalf("expected FieldError for PORT, got: %v", err)
	}
}
