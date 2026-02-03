package rag

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestGetCacheDir(t *testing.T) {
	// Save original env var
	originalEnv := os.Getenv(CacheDirEnv)
	defer os.Setenv(CacheDirEnv, originalEnv)

	tests := []struct {
		name    string
		envVar  string
		wantErr bool
	}{
		{
			name:    "default directory",
			envVar:  "",
			wantErr: false,
		},
		{
			name:    "custom directory from env",
			envVar:  "/custom/cache/dir",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				os.Setenv(CacheDirEnv, tt.envVar)
			} else {
				os.Unsetenv(CacheDirEnv)
			}

			cacheDir, err := GetCacheDir()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCacheDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.envVar != "" {
					if cacheDir != tt.envVar {
						t.Errorf("GetCacheDir() = %v, want %v", cacheDir, tt.envVar)
					}
				} else {
					// Check platform-specific default
					switch runtime.GOOS {
					case "darwin":
						if !filepath.IsAbs(cacheDir) {
							t.Errorf("GetCacheDir() returned relative path: %v", cacheDir)
						}
						if !contains(cacheDir, "Library/Caches") {
							t.Errorf("GetCacheDir() on macOS should contain 'Library/Caches', got: %v", cacheDir)
						}
					case "windows":
						if !filepath.IsAbs(cacheDir) {
							t.Errorf("GetCacheDir() returned relative path: %v", cacheDir)
						}
					default:
						if !filepath.IsAbs(cacheDir) {
							t.Errorf("GetCacheDir() returned relative path: %v", cacheDir)
						}
						if !contains(cacheDir, ".cache") && !contains(cacheDir, "cache") {
							t.Errorf("GetCacheDir() on Linux should contain 'cache', got: %v", cacheDir)
						}
					}
				}
			}
		})
	}
}

func TestHashURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "simple URL",
			url:  "https://example.com/doc.txt",
			want: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
		},
		{
			name: "URL with query params",
			url:  "https://example.com/doc.txt?version=1",
			want: "8c7dd922ad47494fc02c388e12c00eac278d3f4e7c3e3e3e3e3e3e3e3e3e3e3e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hashURL(tt.url)
			if len(got) != 64 {
				t.Errorf("hashURL() returned hash of length %d, want 64", len(got))
			}
			// Hash should be deterministic
			got2 := hashURL(tt.url)
			if got != got2 {
				t.Errorf("hashURL() not deterministic: %v != %v", got, got2)
			}
		})
	}
}

func TestSaveAndLoadMetadata(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	metaPath := filepath.Join(tmpDir, "test.meta.json")

	// Create test metadata
	meta := &CacheMetadata{
		URL:          "https://example.com/doc.txt",
		DownloadedAt: time.Now().Truncate(time.Second),
		ContentHash:  "abc123",
		ETag:         "etag123",
		LastModified: "Mon, 01 Jan 2024 00:00:00 GMT",
		ContentType:  "text/plain",
		Size:         1024,
	}

	// Save metadata
	if err := SaveMetadata(metaPath, meta); err != nil {
		t.Fatalf("SaveMetadata() error = %v", err)
	}

	// Load metadata
	loaded, err := LoadMetadata(metaPath)
	if err != nil {
		t.Fatalf("LoadMetadata() error = %v", err)
	}

	// Compare
	if loaded.URL != meta.URL {
		t.Errorf("URL mismatch: got %v, want %v", loaded.URL, meta.URL)
	}
	if loaded.ContentHash != meta.ContentHash {
		t.Errorf("ContentHash mismatch: got %v, want %v", loaded.ContentHash, meta.ContentHash)
	}
	if loaded.ETag != meta.ETag {
		t.Errorf("ETag mismatch: got %v, want %v", loaded.ETag, meta.ETag)
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    "/home/user/docs/file.txt",
			wantErr: false,
		},
		{
			name:    "path traversal",
			path:    "/home/user/../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "relative path with traversal",
			path:    "docs/../../etc/passwd",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
