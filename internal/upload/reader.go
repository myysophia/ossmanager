package upload

import "io"

// Reader wraps an io.Reader and reports progress to Manager.
type Reader struct {
	r        io.Reader
	id       string
	read     int64
	callback func(int64)
}

func NewReader(id string, r io.Reader) *Reader {
	return &Reader{r: r, id: id}
}

func NewReaderWithCallback(id string, r io.Reader, cb func(int64)) *Reader {
	return &Reader{r: r, id: id, callback: cb}
}

func (pr *Reader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 {
		pr.read += int64(n)
		DefaultManager.Update(pr.id, pr.read)
		if pr.callback != nil {
			pr.callback(pr.read)
		}
	}
	return n, err
}

func (pr *Reader) BytesRead() int64 {
	return pr.read
}
