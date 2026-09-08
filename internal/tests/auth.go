// Package tests contains integration test helpers.
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

const (
	testUser     = "testadmin"
	testEmail    = "testadmin@example.com"
	testPassword = "testpass123"
)

// CreateSuperUser creates the test superuser via podman-compose.
func CreateSuperUser() error {
	cmd := execCommand("podman-compose", "-f", "docker-compose.test.yml", "-p", composeProject, "exec", "-T", "web_recipes", "/opt/recipes/venv/bin/python", "manage.py", "shell", "-c", "from django.contrib.auth import get_user_model; User=get_user_model(); User.objects.filter(username='testadmin').delete(); User.objects.create_superuser('testadmin','testadmin@example.com','testpass123'); print('ok')")
	out, err := runCmd(cmd)
	if err != nil {
		fmt.Printf("CreateSuperUser out: %s\nerr: %v\n", out, err)
	}
	return err
}

// SetupTestData sets up test space and user.
func SetupTestData() error {
	// Base64-encoded Python script to avoid shell quoting issues
	scriptB64 := "ZnJvbSBjb29rYm9vay5tb2RlbHMgaW1wb3J0IFNwYWNlLCBVc2VyU3BhY2UsIEhvdXNlaG9sZCwgSW52ZW50b3J5TG9jYXRpb24KZnJvbSBkamFuZ29fc2NvcGVzIGltcG9ydCBzY29wZXNfZGlzYWJsZWQKZnJvbSBkamFuZ28uY29udHJpYi5hdXRoLm1vZGVscyBpbXBvcnQgR3JvdXAKZnJvbSBkamFuZ28uY29udHJpYi5hdXRoIGltcG9ydCBnZXRfdXNlcl9tb2RlbAoKVXNlciA9IGdldF91c2VyX21vZGVsKCkKdSA9IFVzZXIub2JqZWN0cy5nZXQodXNlcm5hbWU9J3Rlc3RhZG1pbicpCgp3aXRoIHNjb3Blc19kaXNhYmxlZCgpOgogICAgcywgXyA9IFNwYWNlLm9iamVjdHMuZ2V0X29yX2NyZWF0ZShuYW1lPSdUZXN0U3BhY2UnLCBkZWZhdWx0cz17J2NyZWF0ZWRfYnknOiB1fSkKICAgIHMuY3JlYXRlZF9ieSA9IHUKICAgIHMuc2F2ZSgpCiAgICBoLCBfID0gSG91c2Vob2xkLm9iamVjdHMuZ2V0X29yX2NyZWF0ZShuYW1lPSdUZXN0IEhvdXNlaG9sZCcsIHNwYWNlPXMpCiAgICB1cywgXyA9IFVzZXJTcGFjZS5vYmplY3RzLmdldF9vcl9jcmVhdGUodXNlcj11LCBzcGFjZT1zLCBkZWZhdWx0cz17J2FjdGl2ZSc6IFRydWV9KQogICAgdXMuYWN0aXZlID0gVHJ1ZQogICAgdXMuaG91c2Vob2xkID0gaAogICAgdXMuc2F2ZSgpCiAgICBpbCwgXyA9IEludmVudG9yeUxvY2F0aW9uLm9iamVjdHMuZ2V0X29yX2NyZWF0ZSgKICAgICAgICBuYW1lPSdUZXN0IExvY2F0aW9uJywKICAgICAgICBob3VzZWhvbGQ9aCwKICAgICAgICBkZWZhdWx0cz17J2lzX2ZyZWV6ZXInOiBGYWxzZSwgJ2NyZWF0ZWRfYnknOiB1LCAnc3BhY2UnOiBzfQogICAgKQogICAgZ191c2VyID0gR3JvdXAub2JqZWN0cy5nZXQobmFtZT0ndXNlcicpCiAgICB1cy5ncm91cHMuYWRkKGdfdXNlcikKICAgIGdfYWRtaW4gPSBHcm91cC5vYmplY3RzLmdldChuYW1lPSdhZG1pbicpCiAgICB1cy5ncm91cHMuYWRkKGdfYWRtaW4pCiAgICBwcmludCgnc2V0dXAgZG9uZScpCg=="
	cmd := execCommand("podman-compose", "-f", "docker-compose.test.yml", "-p", composeProject, "exec", "-T", "web_recipes", "sh", "-c", "echo '"+scriptB64+"' | base64 -d > /tmp/setup_test_data.py && /opt/recipes/venv/bin/python /opt/recipes/manage.py shell < /tmp/setup_test_data.py")
	out, err := runCmd(cmd)
	if err != nil {
		fmt.Printf("SetupTestData out: %s\nerr: %v\n", out, err)
	}
	return err
}

// GetAPIToken fetches an API token for the test user.
func GetAPIToken(baseURL string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	payload := map[string]string{
		"username": testUser,
		"password": testPassword,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/api-token-auth/", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api-token-auth returned %d", resp.StatusCode)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Token, nil
}

func execCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Dir = composeDir
	return cmd
}
