package craurl

import (
	"bufio"
	"context"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/require"
)

const (
	testData         = "./testdata/urls.txt"
	manyURLsTestData = "./testdata/many-urls.txt"
	badTestData      = "./testdata/bad-url.txt"
	tooLongData      = "./testdata/too-long.txt"
)

func initTestData(source string) (map[string]int, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	urlMap := make(map[string]int)
	file, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		r := rand.Intn(4)
		var statusCode int
		switch r {
		case 0:
			statusCode = http.StatusOK
		case 1:
			statusCode = http.StatusMovedPermanently
		case 2:
			statusCode = http.StatusNotFound
		case 3:
			statusCode = http.StatusInternalServerError
		}
		urlMap[scanner.Text()] = statusCode
	}
	return urlMap, scanner.Err()
}

func TestCrawl(t *testing.T) {
	require := require.New(t)

	testcases := []struct {
		name         string
		data         string
		setup        func()
		teardown     func()
		expectsError bool
	}{
		{
			name:         "token longer that bufio.MaxScanTokenSize error",
			data:         tooLongData,
			expectsError: true,
		},
		{
			name: "bad request method error",
			data: testData,
			setup: func() {
				requestMethod = "@"
			},
			teardown: func() {
				requestMethod = "GET"
			},
			expectsError: true,
		},
		{
			name:         "invalid url error",
			data:         badTestData,
			expectsError: true,
		},
		{
			name: "empty file source ok",
			data: os.DevNull,
		},
		{
			name: "valid data ok",
			data: testData,
			setup: func() {
				urls, err := initTestData(testData)
				require.NoError(err)

				// init mocked responses
				for url, status := range urls {
					gock.New(url).Get("/").Reply(status)
				}
			},
			teardown: func() { gock.OffAll() },
		},
		{
			name: "more than one goroutines loop ok",
			data: manyURLsTestData,
			setup: func() {
				urls, err := initTestData(manyURLsTestData)
				require.NoError(err)

				// init mocked responses
				for url, status := range urls {
					gock.New(url).Get("/").Reply(status)
				}
			},
			teardown: func() { gock.OffAll() },
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}
			if tc.teardown != nil {
				defer tc.teardown()
			}
			file, err := os.Open(tc.data)
			require.NoError(err)
			defer file.Close()
			c, err := New(file)
			require.NoError(err)

			err = c.Crawl(context.Background())

			if tc.expectsError {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}
