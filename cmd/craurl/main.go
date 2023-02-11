package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/kind84/craurl"
)

func main() {
	// sanity checks:
	// one argument must be passed
	usage := `Usage: craurl [filepath]
 filepath	path of the file containing the urls to crawl`
	if len(os.Args) != 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Println(usage)
		os.Exit(0)
	}

	// must be a valid file path
	fp := os.Args[1]
	if _, err := os.Stat(fp); err != nil {
		log.Fatal("file not found")
	}

	// init
	c, err := craurl.NewFromFile(fp)
	if err != nil {
		log.Fatal(err)
	}
	// defer closing the crawler
	defer c.Teardown()

	ctx := context.Background()

	if err := c.Crawl(ctx); err != nil {
		log.Fatal(err)
	}
}
