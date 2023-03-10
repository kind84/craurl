package crawler

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/kind84/craurl/log"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testData                    = "../testdata/urls.txt"
	manyURLsTestData            = "../testdata/many-urls.txt"
	manyURLsWithABadURLTestData = "../testdata/many-urls-with-a-bad-url.txt"
	badTestData                 = "../testdata/bad-url.txt"
	tooLongData                 = "../testdata/too-long.txt"
)

func TestMain(m *testing.M) {
	log.Init()
	code := m.Run()
	os.Exit(code)
}

func initTestData(source string, urls *map[string]int) error {
	rand.Seed(time.Now().UTC().UnixNano())
	urlMap := *urls
	file, err := os.Open(source)
	if err != nil {
		return err
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
	return scanner.Err()
}

func TestCrawl(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	errBanana := errors.New("banana")

	testcases := []struct {
		name          string
		data          string
		setup         func(cancel context.CancelFunc, m *mockStorer, expectedURLs, actualURLs *map[string]int)
		teardown      func()
		expectedError string
	}{
		{
			name:          "token longer that bufio.MaxScanTokenSize error",
			data:          tooLongData,
			expectedError: bufio.ErrTooLong.Error(),
		},
		{
			name: "bad request method error",
			data: testData,
			setup: func(cancel context.CancelFunc, m *mockStorer, expectedURLs, actualURLs *map[string]int) {
				requestMethod = "@"
			},
			teardown: func() {
				requestMethod = "GET"
			},
			expectedError: fmt.Errorf("net/http: invalid method %q", "@").Error(),
		},
		{
			name:          "invalid url error",
			data:          badTestData,
			expectedError: "Get \"pippo.pluto.paperino\": unsupported protocol scheme \"\"",
		},
		{
			name: "empty file source ok",
			data: os.DevNull,
		},
		{
			name: "early storer error",
			data: testData,
			setup: func(cancel context.CancelFunc, m *mockStorer, expectedURLs, actualURLs *map[string]int) {
				err := initTestData(testData, expectedURLs)
				require.NoError(err)

				for url, status := range *expectedURLs {
					// mock network
					gock.New(url).Get("/").Reply(status)
				}

				// mock storer error on the first call
				m.On("StoreResponse", mock.Anything).Return(errBanana).Once()
			},
			teardown:      func() { gock.OffAll() },
			expectedError: errBanana.Error(),
		},
		{
			name: "late storer error",
			data: testData,
			setup: func(cancel context.CancelFunc, m *mockStorer, expectedURLs, actualURLs *map[string]int) {
				err := initTestData(testData, expectedURLs)
				require.NoError(err)

				for url, status := range *expectedURLs {
					// mock network
					gock.New(url).Get("/").Reply(status)
				}

				// mock storer error on the last call
				twoGoodCalls := m.On("StoreResponse", mock.Anything).Return(nil).Twice()
				m.On("StoreResponse", mock.Anything).Return(errBanana).Once().NotBefore(twoGoodCalls)
			},
			teardown:      func() { gock.OffAll() },
			expectedError: errBanana.Error(),
		},
		{
			name: "error on the first goroutine loop",
			data: manyURLsWithABadURLTestData,
			setup: func(cancel context.CancelFunc, m *mockStorer, expectedURLs, actualURLs *map[string]int) {
				err := initTestData(manyURLsWithABadURLTestData, expectedURLs)
				require.NoError(err)
			},
			expectedError: "Get \"pippo.pluto.paperino\": unsupported protocol scheme \"\"",
		},
		{
			name: "context canceled error",
			data: testData,
			setup: func(cancel context.CancelFunc, m *mockStorer, expectedURLs, actualURLs *map[string]int) {
				err := initTestData(testData, expectedURLs)
				require.NoError(err)

				for url, status := range *expectedURLs {
					// mock network
					gock.New(url).Get("/").Reply(status)
				}

				// mock context canceled
				m.On("StoreResponse", mock.Anything).Run(func(args mock.Arguments) {
					cancel()
				}).Return(nil).Once()
			},
			teardown:      func() { gock.OffAll() },
			expectedError: "context canceled",
		},
		{
			name: "valid data ok",
			data: testData,
			setup: func(cancel context.CancelFunc, m *mockStorer, expectedURLs, actualURLs *map[string]int) {
				actURLs := *actualURLs
				err := initTestData(testData, expectedURLs)
				require.NoError(err)

				for url, status := range *expectedURLs {
					// mock network
					gock.New(url).Get("/").Reply(status)

					// mock storer
					m.On("StoreResponse", mock.Anything).Run(func(args mock.Arguments) {
						res := args[0].(Response)

						// The calls to the storer may be out of order and the mock libraty is not
						// smart enough to detect that.  Hence here responses are stored in a map and
						// compared with the expected values at the end of the test
						actURLs[res.URL] = res.Status
					}).Return(nil)
				}
			},
			teardown: func() { gock.OffAll() },
		},
		{
			name: "more than one goroutines loop ok",
			data: manyURLsTestData,
			setup: func(cancel context.CancelFunc, m *mockStorer, expectedURLs, actualURLs *map[string]int) {
				actURLs := *actualURLs
				err := initTestData(manyURLsTestData, expectedURLs)
				require.NoError(err)

				for url, status := range *expectedURLs {
					// mock network
					gock.New(url).Get("/").Reply(status)

					// mock storer
					m.On("StoreResponse", mock.Anything).Run(func(args mock.Arguments) {
						res := args[0].(Response)

						// The calls to the storer may be out of order and the mock libraty is not
						// smart enough to detect that.  Hence here responses are stored in a map and
						// compared with the expected values at the end of the test
						actURLs[res.URL] = res.Status
					}).Return(nil)
				}
			},
			teardown: func() { gock.OffAll() },
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			storer := newMockStorer(t)
			expectedURLs := make(map[string]int)
			actualURLs := make(map[string]int)
			if tc.setup != nil {
				tc.setup(cancel, storer, &expectedURLs, &actualURLs)
			}
			if tc.teardown != nil {
				defer tc.teardown()
			}
			file, err := os.Open(tc.data)
			require.NoError(err)
			defer file.Close()
			c, err := New(file, storer)
			require.NoError(err)

			err = c.Crawl(ctx)

			if tc.expectedError != "" {
				require.ErrorContains(err, tc.expectedError)
				return
			}
			require.NoError(err)
			if assert.Len(actualURLs, len(expectedURLs)) {
				for url, status := range expectedURLs {
					assert.Equal(status, actualURLs[url])
				}
			}
		})
	}
}
