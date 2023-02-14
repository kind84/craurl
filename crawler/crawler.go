package crawler

import (
	"bufio"
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

const maxGoroutines = 100

var requestMethod = "GET"

type storer interface {
	StoreResponse(res Response) error
}

// Response holds response info.
type Response struct {
	URL       string
	Status    int
	Timestamp time.Time
}

// Crawler represents a url crawler.
type Crawler struct {
	source io.Reader
	storer storer
	done   chan struct{}
}

// New retrurn a new instance of Crawler.
func New(source io.Reader, storer storer) (*Crawler, error) {
	c := Crawler{
		source: source,
		storer: storer,
		done:   make(chan struct{}),
	}
	return &c, nil
}

// Crawl holds the main logic of a Craurl. It reads urls from source and calls
// them in parallel. Then it waits for the response and it traces its http
// status.
func (c *Crawler) Crawl(ctx context.Context) error {
	scanner := bufio.NewScanner(c.source)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	resChan := make(chan Response)
	errChan := make(chan error)

	go c.storeResponses(ctx, resChan, errChan)

	// keep looping until source is fully consumed or an error occurs
	for {
		g, ctx := errgroup.WithContext(ctx)

		for i := 0; i < maxGoroutines; i++ {
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return err
				}

				// wait to finish pending requests
				if err := g.Wait(); err != nil {
					close(resChan)
					return err
				}

				// wait for the output to be written
				close(resChan)
				select {
				case err := <-errChan:
					return err
				case <-c.done:
					log.Println("done")
					return nil
				}
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

				res := Response{
					URL:       url,
					Status:    status,
					Timestamp: timestamp,
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case err := <-errChan:
					return err
				case resChan <- res:
					return nil
				}
			})
		}
		if err := g.Wait(); err != nil {
			close(resChan)
			return err
		}
	}
}

// storeResponses waits for responses on the provided channel and stores them
// using the storer.
func (c *Crawler) storeResponses(ctx context.Context, resChan chan Response, errChan chan error) {
	defer close(c.done)

	for res := range resChan {
		err := c.storer.StoreResponse(res)
		if err != nil {
			log.Print(err)
			select {
			case <-ctx.Done():
			case errChan <- err:
			}
			return
		}
	}
}
