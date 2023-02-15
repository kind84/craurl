package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kind84/craurl/crawler"
	"github.com/kind84/craurl/log"
	"github.com/kind84/craurl/storer"
)

func main() {
	log.Init()
	defer log.Sync()

	usage := `Usage: craurl [filepath]
	filepath	path of the file containing the urls to crawl`

	// sanity checks:
	// - one argument must be passed
	if len(os.Args) != 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Println(usage)
		os.Exit(0)
	}
	// - must be a valid file path
	fp := os.Args[1]
	if _, err := os.Stat(fp); err != nil {
		log.Fatalf("file not found: %s", fp)
	}

	// open the source file for reading
	source, err := os.Open(fp)
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()

	// create output file in the same directory of the source file
	dir := filepath.Dir(fp)
	outputPath := filepath.Join(dir, "out.txt")
	output, err := os.Create(outputPath)
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
	log.Infof("output stored in %s", outputPath)
}
