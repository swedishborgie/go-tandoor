package e2e

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

// SetupTestData sets up test space, user, and seed data.
func SetupTestData() error {
	scriptB64 := "ZnJvbSBkamFuZ29fc2NvcGVzIGltcG9ydCBzY29wZXNfZGlzYWJsZWQKZnJvbSBjb29rYm9vay5tb2RlbHMgaW1wb3J0IFNwYWNlLCBVc2VyU3BhY2UsIEhvdXNlaG9sZCwgSW52ZW50b3J5TG9jYXRpb24KZnJvbSBkamFuZ28uY29udHJpYi5hdXRoLm1vZGVscyBpbXBvcnQgR3JvdXAKZnJvbSBkamFuZ28uY29udHJpYi5hdXRoIGltcG9ydCBnZXRfdXNlcl9tb2RlbApVc2VyPWdldF91c2VyX21vZGVsKCkKdT1Vc2VyLm9iamVjdHMuZ2V0KHVzZXJuYW1lPSJ0ZXN0YWRtaW4iKQp3aXRoIHNjb3Blc19kaXNhYmxlZCgpOgogICAgcyxfPVNwYWNlLm9iamVjdHMuZ2V0X29yX2NyZWF0ZShuYW1lPSJUZXN0U3BhY2UiLCBkZWZhdWx0cz17ImNyZWF0ZWRfYnkiOnV9KQogICAgcy5jcmVhdGVkX2J5PXUKICAgIHMuc2F2ZSgpCiAgICBoLF89SG91c2Vob2xkLm9iamVjdHMuZ2V0X29yX2NyZWF0ZShuYW1lPSJUZXN0IEhvdXNlaG9sZCIsIHNwYWNlPXMpCiAgICB1cyxfPVVzZXJTcGFjZS5vYmplY3RzLmdldF9vcl9jcmVhdGUodXNlcj11LCBzcGFjZT1zLCBkZWZhdWx0cz17ImFjdGl2ZSI6VHJ1ZX0pCiAgICB1cy5hY3RpdmU9VHJ1ZQogICAgdXMuaG91c2Vob2xkPWgKICAgIHVzLnNhdmUoKQogICAgaWwsXz1JbnZlbnRvcnlMb2NhdGlvbi5vYmplY3RzLmdldF9vcl9jcmVhdGUobmFtZT0iVGVzdCBMb2NhdGlvbiIsIGhvdXNlaG9sZD1oLCBkZWZhdWx0cz17ImlzX2ZyZWV6ZXIiOkZhbHNlLCJjcmVhdGVkX2J5Ijp1LCJzcGFjZSI6c30pCiAgICBnX3VzZXI9R3JvdXAub2JqZWN0cy5nZXQobmFtZT0idXNlciIpCiAgICB1cy5ncm91cHMuYWRkKGdfdXNlcikKICAgIGdfYWRtaW49R3JvdXAub2JqZWN0cy5nZXQobmFtZT0iYWRtaW4iKQogICAgdXMuZ3JvdXBzLmFkZChnX2FkbWluKQogICAgIyBzZWVkIG1lYWwgdHlwZQogICAgZnJvbSBjb29rYm9vay5tb2RlbHMgaW1wb3J0IE1lYWxUeXBlCiAgICBtdCxfPU1lYWxUeXBlLm9iamVjdHMuZ2V0X29yX2NyZWF0ZShuYW1lPSJCcmVha2Zhc3QiLCBzcGFjZT1zLCBkZWZhdWx0cz17ImNyZWF0ZWRfYnkiOnV9KQogICAgbXQuY3JlYXRlZF9ieT11CiAgICBtdC5zYXZlKCkKICAgICMgc2VlZCByZWNpcGUgYm9vawogICAgZnJvbSBjb29rYm9vay5tb2RlbHMgaW1wb3J0IFJlY2lwZUJvb2sKICAgIHJiLF89UmVjaXBlQm9vay5vYmplY3RzLmdldF9vcl9jcmVhdGUobmFtZT0iVGVzdCBCb29rIiwgc3BhY2U9cywgZGVmYXVsdHM9eyJjcmVhdGVkX2J5Ijp1fSkKICAgIHJiLmNyZWF0ZWRfYnk9dQogICAgcmIuc2F2ZSgpCnByaW50KCJzZXR1cCBkb25lIikK"
	cmd := execCommand("podman-compose", "-f", "docker-compose.test.yml", "-p", composeProject, "exec", "-T", "web_recipes", "sh", "-c", "echo '"+scriptB64+"' | base64 -d > /tmp/setup_test_data.py && /opt/recipes/venv/bin/python /opt/recipes/manage.py shell < /tmp/setup_test_data.py")
	out, err := runCmd(cmd)
	if err != nil {
		fmt.Printf("SetupTestData out: %s\nerr: %v\n", out, err)
	}
	return err
}

// GetAPIToken logs in via the API and returns an access token.
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
