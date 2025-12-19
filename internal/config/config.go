package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// Config captures settings for fetching JWKS documents.
type Config struct {
	CacheFile      string         `yaml:"cache_file"`
	CacheDir       string         `yaml:"cache_dir"`
	TTL            Duration       `yaml:"ttl"`
	RequestTimeout Duration       `yaml:"request_timeout"`
	MaxParallel    int            `yaml:"max_parallel"`
	MaxPerHost     int            `yaml:"max_parallel_per_host"`
	Issuers        []IssuerConfig `yaml:"issuers"`
}

// IssuerConfig represents a single issuer to fetch.
type IssuerConfig struct {
	Issuer  string `yaml:"issuer"`
	JWKSURI string `yaml:"jwks_uri"`
}

// Duration wraps time.Duration for YAML parsing.
type Duration struct {
	time.Duration
}

// UnmarshalYAML implements yaml unmarshalling for duration strings.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var raw string
	if err := value.Decode(&raw); err != nil {
		return err
	}
	dur, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("parse duration: %w", err)
	}
	d.Duration = dur
	return nil
}

// MarshalYAML emits the duration string form.
func (d Duration) MarshalYAML() (any, error) {
	return d.String(), nil
}

// Load reads config from optional file and directory, merging entries.
// Later sources override earlier ones for scalar fields; issuers are appended.
func Load(filePath string, dirPath string) (Config, error) {
	var cfg Config

	if filePath != "" {
		if err := mergeFile(&cfg, filePath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return Config{}, err
			}
		}
	}

	if dirPath != "" {
		if err := mergeDir(&cfg, dirPath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return Config{}, err
			}
		}
	}

	return cfg, nil
}

func mergeDir(cfg *Config, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) == ".yaml" || filepath.Ext(name) == ".yml" {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)
	for _, file := range files {
		if err := mergeFile(cfg, file); err != nil {
			return fmt.Errorf("parse %s: %w", file, err)
		}
	}
	return nil
}

func mergeFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var parsed Config
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return err
	}
	merge(cfg, parsed)
	return nil
}

func merge(dst *Config, src Config) {
	if src.CacheFile != "" {
		dst.CacheFile = src.CacheFile
	}
	if src.CacheDir != "" {
		dst.CacheDir = src.CacheDir
	}
	if src.TTL.Duration != 0 {
		dst.TTL = src.TTL
	}
	if src.RequestTimeout.Duration != 0 {
		dst.RequestTimeout = src.RequestTimeout
	}
	if src.MaxParallel != 0 {
		dst.MaxParallel = src.MaxParallel
	}
	if src.MaxPerHost != 0 {
		dst.MaxPerHost = src.MaxPerHost
	}
	if len(src.Issuers) > 0 {
		dst.Issuers = append(dst.Issuers, src.Issuers...)
	}
}

// Validate ensures required fields exist.
func (c Config) Validate() error {
	if len(c.Issuers) == 0 {
		return errors.New("no issuers configured")
	}
	for i, iss := range c.Issuers {
		if iss.Issuer == "" {
			return fmt.Errorf("issuer[%d] missing issuer", i)
		}
	}
	return nil
}

// Default fills unset fields with sensible defaults.
func (c Config) Defaulted() Config {
	if c.CacheDir == "" {
		c.CacheDir = "/var/cache/jwks"
	}
	if c.TTL.Duration == 0 {
		c.TTL = Duration{Duration: 6 * time.Hour}
	}
	if c.RequestTimeout.Duration == 0 {
		c.RequestTimeout = Duration{Duration: 10 * time.Second}
	}
	return c
}

// Save writes a sample config to the provided path.
func Save(path string, cfg Config, perm fs.FileMode) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}
