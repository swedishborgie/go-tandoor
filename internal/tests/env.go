// Package tests contains integration test helpers.
package tests

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"text/template"
)

type envVars struct {
	Secret string
	DbPass string
}

func generateSecret(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// WriteEnvFile writes the test environment file.
func WriteEnvFile() error {
	secret := generateSecret(32)
	dbPass := generateSecret(16)

	b, err := os.ReadFile(filepath.Join(composeDir, ".env.test.tmpl"))
	if err != nil {
		return err
	}
	tmpl, err := template.New("env").Parse(string(b))
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(composeDir, ".env.test"))
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, envVars{Secret: secret, DbPass: dbPass})
}
