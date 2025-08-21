package providerconfig

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildMinioURL(t *testing.T) {
	tests := []struct {
		name      string
		server    string
		useSSL    bool
		expected  string
		expectErr bool
	}{
		{
			name:     "HTTP with host and port",
			server:   "localhost:9000",
			useSSL:   false,
			expected: "http://localhost:9000",
		},
		{
			name:     "HTTPS with host and port",
			server:   "localhost:9000",
			useSSL:   true,
			expected: "https://localhost:9000",
		},
		{
			name:     "HTTP with host only",
			server:   "minio.example.com",
			useSSL:   false,
			expected: "http://minio.example.com",
		},
		{
			name:     "HTTPS with host only",
			server:   "minio.example.com",
			useSSL:   true,
			expected: "https://minio.example.com",
		},
		{
			name:      "Reject HTTP prefix",
			server:    "http://localhost:9000",
			useSSL:    false,
			expectErr: true,
		},
		{
			name:      "Reject HTTPS prefix",
			server:    "https://localhost:9000",
			useSSL:    true,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildMinioURL(tt.server, tt.useSSL)
			
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("expected %s but got %s", tt.expected, result)
			}
		})
	}
}

func TestValidateMinioCredentials(t *testing.T) {
	// Create test HTTP server
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("MinIO"))
	}))
	defer httpServer.Close()

	// Create test HTTPS server
	httpsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("MinIO"))
	}))
	defer httpsServer.Close()

	// Extract host:port from test servers
	httpServerAddr := httpServer.URL[7:] // Remove "http://" prefix
	httpsServerAddr := httpsServer.URL[8:] // Remove "https://" prefix

	tests := []struct {
		name      string
		creds     map[string]string
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid HTTP credentials",
			creds: map[string]string{
				"minio_server":   httpServerAddr,
				"minio_user":     "testuser",
				"minio_password": "testpass",
				"minio_ssl":      "false",
			},
			expectErr: false,
		},
		{
			name: "Valid HTTPS credentials with insecure",
			creds: map[string]string{
				"minio_server":   httpsServerAddr,
				"minio_user":     "testuser",
				"minio_password": "testpass",
				"minio_ssl":      "true",
				"minio_insecure": "true",
			},
			expectErr: false,
		},
		{
			name: "Invalid SSL value",
			creds: map[string]string{
				"minio_server":   httpServerAddr,
				"minio_user":     "testuser",
				"minio_password": "testpass",
				"minio_ssl":      "maybe",
			},
			expectErr: true,
			errMsg:    "invalid minio_ssl value",
		},
		{
			name: "Invalid insecure value",
			creds: map[string]string{
				"minio_server":    httpServerAddr,
				"minio_user":      "testuser",
				"minio_password":  "testpass",
				"minio_ssl":       "false",
				"minio_insecure":  "maybe",
			},
			expectErr: true,
			errMsg:    "invalid minio_insecure value",
		},
		{
			name: "Server with protocol prefix",
			creds: map[string]string{
				"minio_server":   "http://localhost:9000",
				"minio_user":     "testuser",
				"minio_password": "testpass",
				"minio_ssl":      "false",
			},
			expectErr: true,
			errMsg:    "should not include protocol prefix",
		},
		{
			name: "Unreachable server",
			creds: map[string]string{
				"minio_server":   "nonexistent.localhost:9999",
				"minio_user":     "testuser",
				"minio_password": "testpass",
				"minio_ssl":      "false",
			},
			expectErr: true,
			errMsg:    "failed to connect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMinioCredentials(context.Background(), tt.creds)
			
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !containsString(err.Error(), tt.errMsg) {
					t.Errorf("expected error to contain '%s' but got: %s", tt.errMsg, err.Error())
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateMinioCredentials_DefaultValues(t *testing.T) {
	// Create test HTTP server
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("MinIO"))
	}))
	defer httpServer.Close()

	httpServerAddr := httpServer.URL[7:] // Remove "http://" prefix

	tests := []struct {
		name  string
		creds map[string]string
	}{
		{
			name: "SSL defaults to false when not provided",
			creds: map[string]string{
				"minio_server":   httpServerAddr,
				"minio_user":     "testuser",
				"minio_password": "testpass",
			},
		},
		{
			name: "Insecure defaults to false when not provided",
			creds: map[string]string{
				"minio_server":   httpServerAddr,
				"minio_user":     "testuser",
				"minio_password": "testpass",
				"minio_ssl":      "false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMinioCredentials(context.Background(), tt.creds)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}