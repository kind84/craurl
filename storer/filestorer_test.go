package storer

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kind84/craurl/crawler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileStorer(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	filePath := "./test-new-file-storer.txt"
	defer func() {
		err := os.Remove("./test-new-file-storer.txt")
		require.NoError(err)
	}()
	file, err := os.Create(filePath)
	require.NoError(err)
	defer file.Close()

	fs, err := NewFileStorer(file)

	require.NoError(err)
	require.NoError(file.Sync())
	_, err = file.Seek(0, io.SeekStart)
	require.NoError(err)
	assert.IsType(&os.File{}, fs.output)
	scanner := bufio.NewScanner(fs.output)
	if assert.True(scanner.Scan()) {
		header := scanner.Text()
		expected := fmt.Sprintf("%-35s %s %12s", "URL", "STATUS", "TIMESTAMP")
		assert.Equal(expected, header)
		assert.False(scanner.Scan())
		assert.NoError(scanner.Err())
	}
}

func TestStoreResponse(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	url := "https://google.com"
	status := http.StatusOK
	timestamp := time.Now().UTC()
	filePath := "./test-new-file-storer.txt"
	response := crawler.Response{
		URL:       url,
		Status:    status,
		Timestamp: timestamp,
	}
	defer func() {
		err := os.Remove("./test-new-file-storer.txt")
		require.NoError(err)
	}()
	file, err := os.Create(filePath)
	require.NoError(err)
	defer file.Close()
	fs, err := NewFileStorer(file)
	require.NoError(err)

	err = fs.StoreResponse(response)

	require.NoError(err)
	require.NoError(file.Sync())
	_, err = file.Seek(0, io.SeekStart)
	require.NoError(err)
	assert.IsType(&os.File{}, fs.output)
	scanner := bufio.NewScanner(fs.output)
	assert.True(scanner.Scan()) // scan the header
	if assert.True(scanner.Scan()) {
		header := scanner.Text()
		expected := fmt.Sprintf("%-35s %6d    %s", response.URL, response.Status, response.Timestamp)
		assert.Equal(expected, header)
		assert.False(scanner.Scan())
		assert.NoError(scanner.Err())
	}
}
