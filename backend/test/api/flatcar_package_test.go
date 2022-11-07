package api_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kinvolk/nebraska/backend/pkg/config"
	"github.com/kinvolk/nebraska/backend/pkg/server"
)

func TestHostFlatcarPackage(t *testing.T) {
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	serverPort := uint(6000)
	serverPortStr := fmt.Sprintf(":%d", serverPort)

	conf := &config.Config{
		HostFlatcarPackages: true,
		FlatcarPackagesPath: currentDir,
		AuthMode:            "noop",
		ServerPort:          serverPort,
	}

	db := newDBForTest(t)
	defer db.Close()

	t.Run("file_exists", func(t *testing.T) {
		server, err := server.New(conf, db)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())

		// create a temp file
		fileName := fmt.Sprintf("%s.txt", uuid.NewString())
		file, err := os.Create(path.Join(currentDir, fileName))
		require.NoError(t, err)

		fileString := "This is a test"
		_, err = file.WriteString(fileString)
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		_, err = waitServerReady(fmt.Sprintf("http://localhost:%d", serverPort))
		require.NoError(t, err)

		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/flatcar/%s", serverPort, fileName))
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, fileString, string(bodyBytes))
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// delete the temp file
		err = os.Remove(path.Join(currentDir, fileName))
		require.NoError(t, err)
	})

	t.Run("file_not_exists", func(t *testing.T) {
		server, err := server.New(conf, db)
		require.NotNil(t, server)
		require.NoError(t, err)

		fileName := fmt.Sprintf("%s.txt", uuid.NewString())

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())

		_, err = waitServerReady(fmt.Sprintf("http://localhost:%d", serverPort))
		require.NoError(t, err)

		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/flatcar/%s", serverPort, fileName))
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
