package chanio

import (
	"context"
	"io"
)

// Reader implements chan'ing for an io.Reader object.
type Reader struct {
	output  chan []byte
	rd      io.Reader
	bufsize int
	ctx     context.Context
	err     error
}

// NewReader returns a new Reader
func NewReader(rd io.Reader, size int, ctx context.Context) *Reader {
	r := &Reader{
		output:  make(chan []byte),
		rd:      rd,
		bufsize: size,
		ctx:     ctx,
	}

	r.stopReading()

	return r
}

func (r *Reader) Output() <-chan []byte {
	return r.output
}

func (r *Reader) StartReading() {
	for {
		buf := make([]byte, r.bufsize)

		_, err := r.rd.Read(buf)
		if err != nil {
			r.err = err
			return
		}

		select {
		case r.output <- buf:
		case <-r.ctx.Done():
		}
	}
}

func (r *Reader) GetError() error {
	return r.err
}

func (r *Reader) stopReading() {
	readCloser, ok := r.rd.(io.ReadCloser)
	if !ok {
		return
	}

	go func() {
		<-r.ctx.Done()
		readCloser.Close()
	}()
}
