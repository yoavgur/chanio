package main

import (
	"context"
	"errors"
	"io"
)

const defaultBufSize = 4096

var ErrReadTerminated = errors.New("read terminated")
var ErrWriteTerminated = errors.New("write terminated")

// Reader implements chan'ing for an io.Reader object.
type Reader struct {
	rd      io.Reader
	bufsize int
	err     error
	done    chan struct{}
	ctx     context.Context
	output  chan []byte
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

func NewReaderSizeContext(rd io.Reader, size int, ctx context.Context) *Reader {
	r := &Reader{
		rd:      rd,
		bufsize: size,
		ctx:     ctx,
		done:    make(chan struct{}),
		output:  make(chan []byte),
	}

	r.readLoop()
	r.stopReading()

	return r
}

func (r *Reader) Read() <-chan []byte {
	return r.output
}

func (r *Reader) readLoop() {
	go func() {
		defer close(r.done)
		defer close(r.output)

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
			case r.output <- buf:
			case <-r.ctx.Done():
				return
			}
		}
	}()
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

type Writer struct {
	wr    io.Writer
	err   error
	done  chan struct{}
	ctx   context.Context
	input chan []byte
}

func NewWriterContext(wr io.Writer, ctx context.Context) *Writer {
	return NewWriterSizeContext(wr, defaultBufSize, ctx)
}

func NewWriterSize(wr io.Writer, size int) *Writer {
	return NewWriterSizeContext(wr, size, context.Background())
}

func NewWriter(wr io.Writer) *Writer {
	return NewWriterSizeContext(wr, defaultBufSize, context.Background())
}

func NewWriterSizeContext(wr io.Writer, size int, ctx context.Context) *Writer {
	w := &Writer{
		wr:    wr,
		ctx:   ctx,
		done:  make(chan struct{}),
		input: make(chan []byte),
	}

	w.writeLoop()
	w.stopWriting()

	return w
}

func (w *Writer) Write() chan<- []byte {
	return w.input
}

func (w *Writer) writeLoop() {
	go func() {
		defer close(w.done)

		for {
			select {
			case buf, ok := <-w.input:
				if !ok {
					return
				}
				_, err := w.wr.Write(buf)
				if err != nil {
					w.err = err
				}
			case <-w.ctx.Done():
				return
			}
		}
	}()
}

func (r *Writer) GetError() error {
	err := r.err
	r.err = nil
	return err
}

func (w *Writer) stopWriting() {
	writeCloser, ok := w.wr.(io.WriteCloser)
	if !ok {
		return
	}

	go func() {
		select {
		case <-w.ctx.Done():
			w.err = ErrWriteTerminated
			writeCloser.Close()

		case <-w.done:
		}
	}()
}
