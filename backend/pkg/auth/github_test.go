package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGithubClient(t *testing.T) {
	for _, tt := range []struct {
		name          string
		enterpriseURL string
		wantBaseURL   string
		wantUploadURL string
	}{
		{
			name:          "public github",
			enterpriseURL: "",
			wantBaseURL:   "https://api.github.com/",
			wantUploadURL: "https://uploads.github.com/",
		},
		{
			name:          "enterprise instance",
			enterpriseURL: "https://github.example.com",
			wantBaseURL:   "https://github.example.com/api/v3/",
			wantUploadURL: "https://github.example.com/api/uploads/",
		},
		{
			name:          "enterprise instance with trailing slash",
			enterpriseURL: "https://github.example.com/",
			wantBaseURL:   "https://github.example.com/api/v3/",
			wantUploadURL: "https://github.example.com/api/uploads/",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client, err := newGithubClient(&http.Client{}, tt.enterpriseURL)
			require.NoError(t, err)

			assert.Equal(t, tt.wantBaseURL, client.BaseURL())
			assert.Equal(t, tt.wantUploadURL, client.UploadURL())
		})
	}
}
