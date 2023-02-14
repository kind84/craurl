package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kind84/craurl/crawler"
	"github.com/kind84/craurl/log"
	"github.com/kind84/craurl/storer"
)

func main() {
	log.Init()
	defer log.Sync()

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

	// open the source for reading
	source, err := os.Open(fp)
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()

	// create output file
	output, err := os.Create("out.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	// init file storer
	storer, err := storer.NewFileStorer(output)
	if err != nil {
		log.Fatal(err)
	}

	// init url crawler
	c, err := crawler.New(source, storer)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	if err := c.Crawl(ctx); err != nil {
		log.Fatal(err)
	}
}
