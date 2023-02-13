package craurl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

const maxGoroutines = 100

var requestMethod = "GET"

// Craurl represents a url crawler.
type Craurl struct {
	source io.Reader
}

// New retrurn a new instance of Craurl.
func New(source io.Reader) (*Craurl, error) {
	c := Craurl{source: source}
	return &c, nil
}

// Crawl holds the main logic of a Craurl. It reads urls from source and calls
// them in parallel. Then it waits for the response and it traces its http
// status.
func (c *Craurl) Crawl(ctx context.Context) error {
	scanner := bufio.NewScanner(c.source)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// keep looping until source is fully consumed or an error occurs
	for {
		g, ctx := errgroup.WithContext(ctx)
		var scannerErr error

		for i := 0; i < maxGoroutines; i++ {
			ok := scanner.Scan()
			if !ok {
				scannerErr = scanner.Err()
				if scannerErr == nil {
					// scanner returns false and nil error only in case of EOF
					scannerErr = io.EOF
				}
				break
			}

			url := scanner.Text()

			g.Go(func() error {
				url := url
				log.Printf("calling %s", url)
				req, err := http.NewRequestWithContext(ctx, requestMethod, url, nil)
				if err != nil {
					return err
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				status := resp.StatusCode
				timestamp := time.Now().UTC()
				fmt.Printf("%s   %d  %s\n", url, status, timestamp)

				return nil
			})
		}
		if scannerErr != nil && !errors.Is(scannerErr, io.EOF) {
			return scannerErr
		}
		if err := g.Wait(); err != nil {
			return err
		}
		if errors.Is(scannerErr, io.EOF) {
			log.Println("done")
			return nil
		}
	}
}
