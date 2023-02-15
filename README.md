# Craurl

### URL crawler in Go

Craurl is a minimal URL crawler. It can be used as a standalone binary or imported as a library.

## Usage

Clone the repository:

```sh
$ git clone https://github.com/kind84/craurl.git
```

### Build

```sh
$ make build
```

The `craurl` binary will be created in the `/dist` folder of the project.

### Run

```sh
$ ./craurl <path-to-your-urls-file>
```

The application will output the result in a file called `out.txt` in the same folder of the input file.

## As a Library

You can have a `craurl.Crawler` using the `New` function with this signature:

```go
func New(source io.Reader, storer storer) (*Crawler, error)
```

With this you can use the `Crawler` in your project by passing any `io.Reader` as source:

```go
// create file output
output, err := os.Create("my-output.txt")
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
```

It requires a `storer` that fullfills the `crawler.storer` interface:

```go
type storer interface {
	StoreResponse(res Response) error
}
```

You can use the `storer.FileStorer` or implement your own.

### Docker

To use it as a Docker container:

- **Build** the Docker image:
```sh
$ make docker
```

- **Run** the container:

```sh
$ docker run -v /folder-with-urls-file:/data craurl
```

By default it expects to get `urls.txt` file mounted into the `/data` directory.
You can override this by setting the `$DATA` and `$URLS` environment variables.

### Test

- Run all the test in the project:

```sh
$ make test
```

- Get a test coverage report:

```sh
$ make cover
```

- Generate a html coverage report:

```sh
$ make cover-html
```

This will produce a coverage.html file that you can open in your browser.
