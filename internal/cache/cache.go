package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Entry represents a cached JWKS entry.
type Entry struct {
	Expiration float64        `json:"expiration"`
	NextUpdate float64        `json:"next_update,omitempty"`
	JWKS       map[string]any `json:"jwks"`
	ETag       string         `json:"etag,omitempty"`
}

// Document is a map keyed by issuer.
type Document map[string]Entry

// BuildEntry constructs a cache entry with the provided TTL.
// If useSubsecond is false (default), timestamps are stored as whole seconds only.
func BuildEntry(jwks map[string]any, ttl time.Duration, useSubsecond bool) Entry {
	now := time.Now()
	exp := now.Add(ttl)
	next := now.Add(ttl * 3 / 4) // refresh a bit before expiry

	var expiration, nextUpdate float64
	if useSubsecond {
		expiration = float64(exp.Unix()) + float64(exp.Nanosecond())/1e9
		nextUpdate = float64(next.Unix()) + float64(next.Nanosecond())/1e9
	} else {
		expiration = float64(exp.Unix())
		nextUpdate = float64(next.Unix())
	}

	return Entry{
		Expiration: expiration,
		NextUpdate: nextUpdate,
		JWKS:       jwks,
	}
}

// WriteDirectory writes one file per issuer using the hashed naming scheme.
func WriteDirectory(dir string, doc Document) error {
	if dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	for issuer, entry := range doc {
		name := hashedFileName(issuer)
		path := filepath.Join(dir, name)
		if err := writeFile(path, Document{issuer: entry}); err != nil {
			return err
		}
	}
	return nil
}

// WriteFile writes the full document into a single JSON file.
func WriteFile(path string, doc Document) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	return writeFile(path, doc)
}

func writeFile(path string, doc Document) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	return nil
}

func hashedFileName(issuer string) string {
	sum := sha256.Sum256([]byte(issuer))
	return hex.EncodeToString(sum[:])[:8]
}

// LoadFile parses a cache document from a file that may contain multiple
// concatenated JSON objects. Later objects override earlier ones.
func LoadFile(path string) (Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseConcatJSON(data)
}

// LoadHashed finds the first hashed file matching the issuer in dir.
func LoadHashed(dir, issuer string) (Entry, bool, error) {
	if dir == "" {
		return Entry{}, false, nil
	}

	name := hashedFileName(issuer)
	path := filepath.Join(dir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Entry{}, false, nil
		}
		return Entry{}, false, err
	}
	doc, err := parseConcatJSON(data)
	if err != nil {
		return Entry{}, false, err
	}
	entry, ok := doc[issuer]
	return entry, ok, nil
}

func parseConcatJSON(data []byte) (Document, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	out := make(Document)
	for {
		var obj map[string]Entry
		if err := dec.Decode(&obj); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		for k, v := range obj {
			out[k] = v
		}
	}
	return out, nil
}
