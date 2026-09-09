//go:build integration
// +build integration

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

var (
	baseURL    = "http://localhost:8080"
	token      string
	cliBin     string
	fdcSrv     *http.Server
	fdcBaseURL string
)

func init() {
	println("CLI E2E SUITE INIT")
}

func TestMain(m *testing.M) {
	fmt.Println("CLI TESTMAIN START")
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		os.Exit(m.Run())
	}
	fmt.Println("WriteEnvFile")
	if err := WriteEnvFile(); err != nil {
		panic(fmt.Errorf("WriteEnvFile: %w", err))
	}
	fmt.Println("Start")
	if err := Start(); err != nil {
		panic(fmt.Errorf("Start: %w", err))
	}
	fmt.Println("waitReady")
	if err := waitReady(); err != nil {
		panic(fmt.Errorf("waitReady: %w", err))
	}
	fmt.Println("CreateSuperUser")
	if err := CreateSuperUser(); err != nil {
		panic(fmt.Errorf("CreateSuperUser: %w", err))
	}
	fmt.Println("SetupTestData")
	if err := SetupTestData(); err != nil {
		panic(fmt.Errorf("SetupTestData: %w", err))
	}
	fmt.Println("GetAPIToken")
	var err error
	token, err = GetAPIToken(baseURL)
	if err != nil {
		panic(fmt.Errorf("GetAPIToken: %w", err))
	}
	fmt.Println("Start mock FDC")
	fdcBaseURL, fdcSrv = startMockFDCSrv()
	fmt.Println("Build CLI")
	cliBin, err = buildCLI()
	if err != nil {
		panic(fmt.Errorf("build CLI: %w", err))
	}
	code := m.Run()
	if fdcSrv != nil {
		fdcSrv.Shutdown(context.Background())
	}
	_ = Stop()
	os.Exit(code)
}

func waitReady() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	client := &http.Client{Timeout: 2 * time.Second}
	url := baseURL + "/"
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for tandoor")
		default:
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 400 {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
}

func buildCLI() (string, error) {
	tmp := filepath.Join(os.TempDir(), "tandoor-e2e")
	cmd := exec.Command("go", "build", "-o", tmp, "./cmd/tandoor")
	cmd.Dir = repoRoot()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("go build failed: %v\n%s", err, out)
	}
	return tmp, nil
}

// repoRoot returns the repository root, derived from this file's location
// (internal/cli/e2e is three levels below the root).
func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}

func runCLI(args ...string) (stdout string, stderr string, exitCode int, err error) {
	cmd := exec.Command(cliBin, args...)
	env := append(os.Environ(), "TANDOOR_BASE_URL="+baseURL, "TANDOOR_TOKEN="+token, "FDC_API_KEY=dummy")
	if fdcBaseURL != "" {
		env = append(env, "FDC_BASE_URL="+fdcBaseURL)
	}
	cmd.Env = env
	stdoutBytes, err := cmd.CombinedOutput()
	output := string(stdoutBytes)
	// split stdout/stderr? CombinedOutput merges. For simplicity return combined.
	// We'll parse exit code from err.
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	} else {
		exitCode = 0
	}
	return output, "", exitCode, err
}

func startMockFDCSrv() (string, *http.Server) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/foods/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		// Return deterministic mock response
		res := map[string]any{
			"totalHits":   1,
			"currentPage": 0,
			"totalPages":  1,
			"foods": []map[string]any{
				{
					"fdcId":       123456,
					"dataType":    "Foundation",
					"description": query + " mock food",
					"score":       0.99,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})
	mux.HandleFunc("/v1/food/", func(w http.ResponseWriter, r *http.Request) {
		// Simplified: return fixed food with nutrients
		food := map[string]any{
			"fdcId":       123456,
			"description": "Mock Food",
			"dataType":    "Foundation",
			"foodNutrients": []map[string]any{
				{"nutrient": map[string]any{"id": 1008, "name": "Energy", "unitName": "kcal"}, "amount": 52.0},
				{"nutrient": map[string]any{"id": 1003, "name": "Protein", "unitName": "g"}, "amount": 0.3},
				{"nutrient": map[string]any{"id": 1004, "name": "Total lipid", "unitName": "g"}, "amount": 0.1},
				{"nutrient": map[string]any{"id": 1005, "name": "Carbohydrate", "unitName": "g"}, "amount": 14.0},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(food)
	})
	srv := &http.Server{Addr: "127.0.0.1:0", Handler: mux}
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		panic(err)
	}
	go srv.Serve(ln)
	base := "http://" + ln.Addr().String()
	return base, srv
}
