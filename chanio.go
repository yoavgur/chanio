package main

import (
	"context"
	"errors"
	"io"
)

const defaultBufSize = 4096

var ErrReadTerminated = errors.New("read terminated")

// Reader implements chan'ing for an io.Reader object.
type Reader struct {
	rd      io.Reader
	bufsize int
	err     error
	done    chan struct{}
	ctx     context.Context
}

func NewReaderSizeContext(rd io.Reader, size int, ctx context.Context) *Reader {
	r := &Reader{
		rd:      rd,
		bufsize: size,
		ctx:     ctx,
		done:    make(chan struct{}),
	}

	r.stopReading()

	return r
}

func NewReaderContext(rd io.Reader, ctx context.Context) *Reader {
	return NewReaderSizeContext(rd, defaultBufSize, ctx)
}

func NewReaderSize(rd io.Reader, size int) *Reader {
	return NewReaderSizeContext(rd, size, context.Background())
}

func NewReader(rd io.Reader) *Reader {
	return NewReaderSizeContext(rd, defaultBufSize, context.Background())
}

func (r *Reader) Read() <-chan []byte {
	output := make(chan []byte)

	go func() {
		defer close(r.done)
		defer close(output)

		for {
			buf := make([]byte, r.bufsize)

			_, err := r.rd.Read(buf)
			if err != nil {
				select {
				case <-r.ctx.Done():
				default:
					if err != io.EOF {
						r.err = err
					}
				}
				return
			}

			select {
			case output <- buf:
			case <-r.ctx.Done():
				return
			}
		}
	}()

	return output
}

func (r *Reader) GetError() error {
	err := r.err
	r.err = nil
	return err
}

func (r *Reader) stopReading() {
	readCloser, ok := r.rd.(io.ReadCloser)
	if !ok {
		return
	}

	go func() {
		select {
		case <-r.ctx.Done():
			r.err = ErrReadTerminated
			readCloser.Close()

		case <-r.done:
		}
	}()
}
