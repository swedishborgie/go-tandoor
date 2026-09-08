// cmd/tandoor-cli/helpers.go

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/detector"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/property"
)

// propertyTypeIDs returns the ids of property types attached to f.
func propertyTypeIDs(f *food.Food) []int {
	ids := make([]int, 0, len(f.Properties))
	for _, p := range f.Properties {
		ids = append(ids, p.Type.ID)
	}
	return ids
}

// fetchAllPropertyTypes enumerates every property type on the server.
func fetchAllPropertyTypes(ctx context.Context, c *tandoor.Client) ([]property.Type, error) {
	svc := property.NewTypeService(c)
	page, err := svc.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list property types: %w", err)
	}
	return pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[property.Type], error) {
		return svc.List(ctx, &property.TypeListOptions{ListOptions: pagination.ListOptions{Page: pageNum}})
	})
}

// newAuditRegistry returns a detector registry with the built-in detectors
// plus missing_properties, whose expected type set is enumerated from the
// server (GET /api/property-type/). If the enumeration fails, it warns on
// stderr and returns the registry without the detector — the audit's core
// job (names/FDC) still runs. Callers should invoke this once per command
// invocation, not per food.
func newAuditRegistry(ctx context.Context, c *tandoor.Client, cmd *cli.Command) *detector.Registry {
	reg := detector.NewRegistry()
	types, err := fetchAllPropertyTypes(ctx, c)
	if err != nil {
		fmt.Fprintf(cmd.ErrWriter, "[warn] audit: could not enumerate property types from server, skipping missing_properties check: %v\n", err)
		return reg
	}
	expected := make([]detector.PropertyTypeInfo, 0, len(types))
	for _, t := range types {
		expected = append(expected, detector.PropertyTypeInfo{ID: t.ID, Name: t.Name})
	}
	reg.Add(detector.NewMissingPropertiesDetector(expected))
	return reg
}

// oidcLogin initiates an OIDC login flow by opening the browser and
// catching the callback on a local HTTP server.
func oidcLogin(ctx context.Context, c *tandoor.Client, oidcBackend, redirectHost string, redirectPort int) (string, error) {
	hostPort := net.JoinHostPort(redirectHost, strconv.Itoa(redirectPort))
	redirectURL := "http://" + hostPort + "/callback"
	authURL := fmt.Sprintf("%s/accounts/social/login/%s/?next=%%2F&redirect_uri=%s",
		c.BaseURLOrigin(), oidcBackend, url.QueryEscape(redirectURL))

	fmt.Printf("Opening browser for OIDC login...\n")
	fmt.Printf("URL: %s\n", authURL)

	// Start local callback server.
	tokenCh := make(chan string, 1)
	srv := &http.Server{Addr: fmt.Sprintf("%s:%d", redirectHost, redirectPort)}
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("access_token")
		if token == "" {
			_ = r.ParseForm()
			token = r.PostFormValue("access_token")
		}
		if token == "" {
			http.Error(w, "no access_token in callback", http.StatusBadRequest)
			tokenCh <- ""
			return
		}
		fmt.Fprintf(w, "<h1>Login successful!</h1><p>You can close this window.</p>")
		tokenCh <- token
	})
	srv.Handler = mux

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "callback server error: %v\n", err)
		}
	}()

	// Open browser.
	if err := openBrowser(ctx, authURL); err != nil {
		fmt.Fprintf(os.Stderr, "failed to open browser: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please open this URL manually: %s\n", authURL)
	}

	// Wait for callback with timeout.
	select {
	case token := <-tokenCh:
		if err := srv.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warn: close callback server: %v\n", err)
		}
		if token == "" {
			return "", fmt.Errorf("OIDC callback returned empty token")
		}
		return token, nil
	case <-time.After(120 * time.Second):
		if err := srv.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warn: close callback server: %v\n", err)
		}
		return "", fmt.Errorf("OIDC login timed out after 120s")
	}
}

// openBrowser opens the default browser to the given URL.
func openBrowser(ctx context.Context, browserURL string) error {
	var cmd string
	var args []string

	switch os.Getenv("GOOS") {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	default:
		cmd = "xdg-open"
	}

	if len(args) == 0 {
		args = []string{browserURL}
	} else {
		args = append(args, browserURL)
	}

	c := exec.CommandContext(ctx, cmd, args...)
	return c.Start()
}

func parseIDArg(arg string) (int, error) {
	if arg == "" {
		return 0, fmt.Errorf("ID argument is required")
	}
	id, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid ID %q: %w", arg, err)
	}
	return id, nil
}

func printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func applyJQ(ctx context.Context, filter string, data []byte) ([]byte, error) {
	if filter == "" {
		return data, nil
	}
	// Use jq if available
	cmd := exec.CommandContext(ctx, "jq", "-c", filter)
	cmd.Stdin = bytes.NewReader(data)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("jq filter failed: %w", err)
	}
	return out, nil
}

func outputWithJQ(ctx context.Context, v any, filter string, outputFile string) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if filter != "" {
		data, err = applyJQ(ctx, filter, data)
		if err != nil {
			return err
		}
	}
	if outputFile != "" {
		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "[output] wrote %s\n", outputFile)
		return nil
	}
	fmt.Println(string(data))
	return nil
}

func printError(err error) {
	if err == nil {
		return
	}
	// Try to detect TandoorError for structured output
	var tErr *tandoor.TandoorError
	if errors.As(err, &tErr) {
		out := map[string]any{
			"status":  tErr.StatusCode,
			"message": tErr.Message,
		}
		if details := tErr.Details(); len(details) > 0 {
			out["details"] = details
		}
		data, mErr := json.MarshalIndent(out, "", "  ")
		if mErr != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		fmt.Fprintln(os.Stderr, string(data))
		return
	}
	fmt.Fprintln(os.Stderr, err)
}

func parseIntCSV(value string) ([]int, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	ids := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q: %w", part, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func parseDateOrDateTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("date value is required")
	}
	if ts, err := time.Parse(time.RFC3339, value); err == nil {
		return ts, nil
	}
	if ts, err := time.Parse("2006-01-02", value); err == nil {
		return ts, nil
	}
	return time.Time{}, fmt.Errorf("invalid date %q (use RFC3339 or YYYY-MM-DD)", value)
}
