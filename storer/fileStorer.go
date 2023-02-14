package storer

import (
	"fmt"
	"os"

	"github.com/kind84/craurl/crawler"
)

// FileStorer is a Storer that writes to a file.
type FileStorer struct {
	output *os.File
}

// NewFileStorer returns a new instance of FileStorer. The provided file path
// is trusted to be valid. The consumer of the FileStorer is responsible for
// closing the file.
func NewFileStorer(out *os.File) (*FileStorer, error) {
	_, err := out.WriteString(fmt.Sprintf("%-35s %s %12s\n", "URL", "STATUS", "TIMESTAMP"))
	if err != nil {
		return nil, err
	}

	s := FileStorer{output: out}
	return &s, nil
}

// StoreResponse writes the response object into a line of the file.
func (fs *FileStorer) StoreResponse(res crawler.Response) error {
	line := fmt.Sprintf("%-35s %6d    %s\n", res.URL, res.Status, res.Timestamp)
	_, err := fs.output.WriteString(line)
	if err != nil {
		return err
	}
	return nil
}
