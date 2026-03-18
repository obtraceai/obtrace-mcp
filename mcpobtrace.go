// Package mcpobtrace provides the core MCP server implementation for Obtrace.
//
// It exposes Obtrace's observability platform capabilities (logs, traces, metrics,
// dashboards, incidents, AI analysis) as MCP tools that can be consumed by
// LLM-powered agents and IDE integrations.
package mcpobtrace

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// ObtraceConfig holds the configuration for connecting to an Obtrace instance.
type ObtraceConfig struct {
	// URL is the base URL of the Obtrace API (e.g. https://api.obtrace.ai).
	URL string

	// APIKey is the authentication token for Obtrace API access.
	APIKey string

	// TenantID scopes all operations to a specific tenant.
	TenantID string

	// ProjectID scopes operations to a specific project within the tenant.
	ProjectID string

	// TLS holds optional TLS configuration for mTLS or custom CAs.
	TLS *TLSConfig

	// HTTPClient is an optional pre-configured HTTP client.
	// If nil, a default client with sensible timeouts is created.
	HTTPClient *http.Client
}

// TLSConfig holds TLS-specific configuration.
type TLSConfig struct {
	// CACertPath is the path to a custom CA certificate file.
	CACertPath string

	// ClientCertPath is the path to a client certificate for mTLS.
	ClientCertPath string

	// ClientKeyPath is the path to the client key for mTLS.
	ClientKeyPath string

	// InsecureSkipVerify disables certificate verification (use only for development).
	InsecureSkipVerify bool
}

type contextKey string

const (
	configContextKey contextKey = "obtrace-config"
	clientContextKey contextKey = "obtrace-client"
)

// WithObtraceConfig stores the ObtraceConfig in the context.
func WithObtraceConfig(ctx context.Context, cfg *ObtraceConfig) context.Context {
	return context.WithValue(ctx, configContextKey, cfg)
}

// ObtraceConfigFromContext retrieves the ObtraceConfig from the context.
func ObtraceConfigFromContext(ctx context.Context) (*ObtraceConfig, error) {
	cfg, ok := ctx.Value(configContextKey).(*ObtraceConfig)
	if !ok || cfg == nil {
		return nil, fmt.Errorf("obtrace config not found in context")
	}
	return cfg, nil
}

// WithObtraceClient stores an HTTP client in the context.
func WithObtraceClient(ctx context.Context, client *http.Client) context.Context {
	return context.WithValue(ctx, clientContextKey, client)
}

// ObtraceClientFromContext retrieves the HTTP client from the context,
// falling back to a default client if none is set.
func ObtraceClientFromContext(ctx context.Context) *http.Client {
	client, ok := ctx.Value(clientContextKey).(*http.Client)
	if !ok || client == nil {
		return defaultHTTPClient()
	}
	return client
}

// ConfigFromEnv creates an ObtraceConfig from environment variables.
//
// Environment variables:
//   - OBTRACE_URL: Base URL of the Obtrace API (required)
//   - OBTRACE_API_KEY: API key for authentication (required)
//   - OBTRACE_TENANT_ID: Tenant ID for scoping
//   - OBTRACE_PROJECT_ID: Project ID for scoping
//   - OBTRACE_TLS_CA_CERT: Path to CA certificate
//   - OBTRACE_TLS_CLIENT_CERT: Path to client certificate
//   - OBTRACE_TLS_CLIENT_KEY: Path to client key
//   - OBTRACE_TLS_INSECURE: Set to "true" to skip TLS verification
func ConfigFromEnv() (*ObtraceConfig, error) {
	url := os.Getenv("OBTRACE_URL")
	if url == "" {
		return nil, fmt.Errorf("OBTRACE_URL environment variable is required")
	}
	url = strings.TrimRight(url, "/")

	apiKey := os.Getenv("OBTRACE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OBTRACE_API_KEY environment variable is required")
	}

	cfg := &ObtraceConfig{
		URL:       url,
		APIKey:    apiKey,
		TenantID:  os.Getenv("OBTRACE_TENANT_ID"),
		ProjectID: os.Getenv("OBTRACE_PROJECT_ID"),
	}

	// TLS configuration
	caCert := os.Getenv("OBTRACE_TLS_CA_CERT")
	clientCert := os.Getenv("OBTRACE_TLS_CLIENT_CERT")
	clientKey := os.Getenv("OBTRACE_TLS_CLIENT_KEY")
	insecure := os.Getenv("OBTRACE_TLS_INSECURE")

	if caCert != "" || clientCert != "" || insecure == "true" {
		cfg.TLS = &TLSConfig{
			CACertPath:         caCert,
			ClientCertPath:     clientCert,
			ClientKeyPath:      clientKey,
			InsecureSkipVerify: insecure == "true",
		}
	}

	return cfg, nil
}

// NewObtraceHTTPClient creates an HTTP client configured for Obtrace API access.
func NewObtraceHTTPClient(cfg *ObtraceConfig) (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = 10

	if cfg.TLS != nil {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: cfg.TLS.InsecureSkipVerify, //nolint:gosec // user-configured
		}

		if cfg.TLS.ClientCertPath != "" && cfg.TLS.ClientKeyPath != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLS.ClientCertPath, cfg.TLS.ClientKeyPath)
			if err != nil {
				return nil, fmt.Errorf("loading client certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		transport.TLSClientConfig = tlsConfig
	}

	// Add API key to all requests via a round tripper wrapper.
	client := &http.Client{
		Transport: &authTransport{
			base:   transport,
			apiKey: cfg.APIKey,
		},
		Timeout: 30 * time.Second,
	}

	return client, nil
}

// authTransport injects the API key header into every outgoing request.
type authTransport struct {
	base   http.RoundTripper
	apiKey string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("Authorization", "Bearer "+t.apiKey)
	reqClone.Header.Set("User-Agent", "obtrace-mcp/0.1.0")
	return t.base.RoundTrip(reqClone)
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// ContextFunc is a function that adds values to a context.
// Used to compose context setup for different transport modes.
type ContextFunc func(ctx context.Context) context.Context

// ComposeContextFuncs chains multiple ContextFuncs together.
func ComposeContextFuncs(fns ...ContextFunc) ContextFunc {
	return func(ctx context.Context) context.Context {
		for _, fn := range fns {
			ctx = fn(ctx)
		}
		return ctx
	}
}

// StdioContextFunc returns a ContextFunc that injects config from environment variables.
func StdioContextFunc() (ContextFunc, error) {
	cfg, err := ConfigFromEnv()
	if err != nil {
		return nil, err
	}

	client, err := NewObtraceHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) context.Context {
		ctx = WithObtraceConfig(ctx, cfg)
		ctx = WithObtraceClient(ctx, client)
		return ctx
	}, nil
}
