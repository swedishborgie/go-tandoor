//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/swedishborgie/go-tandoor"
)

func init() {
	println("SUITE INIT")
}

var (
	baseURL = "http://localhost:8080"
	client  *tandoor.Client
)

func strPtr(s string) *string     { return &s }
func floatPtr(f float64) *float64 { return &f }

func TestMain(m *testing.M) {
	fmt.Println("TESTMAIN START")
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
	token, err := GetAPIToken(baseURL)
	if err != nil {
		panic(fmt.Errorf("GetAPIToken: %w", err))
	}
	c, err := tandoor.NewClient(baseURL, tandoor.WithAccessToken(token))
	if err != nil {
		panic(err)
	}
	client = c
	code := m.Run()
	_ = Stop()
	os.Exit(code)
}

func Client() *tandoor.Client {
	return client
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
