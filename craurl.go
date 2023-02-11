package craurl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

const maxGoroutines = 100

// Craurl represents a url crawler.
type Craurl struct {
	reader io.ReadSeekCloser
}

// NewFromFile retrurn a new instance of Craurl that reads from a file sorce.
func NewFromFile(source string) (*Craurl, error) {
	// open the file for reading
	file, err := os.Open(source)
	if err != nil {
		return nil, err
	}

	c := Craurl{reader: file}
	return &c, nil
}

// Teardown gracefully stops the Craurl instance.
func (c *Craurl) Teardown() error {
	// close the file
	log.Print("stopping..")
	return c.reader.Close()
}

// Crawl holds the main logic of a Craurl. It reads urls from source and calls
// them in parallel. Then it waits for the response and it traces its http
// status.
func (c *Craurl) Crawl(ctx context.Context) error {
	scanner := bufio.NewScanner(c.reader)
	g, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var scannerErr error
	for {
		for i := 0; i < maxGoroutines; i++ {
			ok := scanner.Scan()
			if !ok {
				scannerErr = scanner.Err()
				if scannerErr == nil {
					scannerErr = io.EOF
				}
				break
			}
			// parse url
			url := scanner.Text()

			g.Go(func() error {
				log.Printf("calling %s", url)
				// call url
				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				if err != nil {
					return err
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				// read status from response
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
