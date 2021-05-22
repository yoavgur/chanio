package chanio

import (
	"context"
	"errors"
	"io"
)

const defaultBufSize = 4096

var ErrIOTerminated = errors.New("IO terminated")

type IO struct {
	err  error
	done chan struct{}
	ctx  context.Context
}

func newIO(ctx context.Context) *IO {
	return &IO{
		ctx:  ctx,
		done: make(chan struct{}),
	}
}

func (i *IO) GetError() error {
	err := i.err
	i.err = nil
	return err
}

func (i *IO) stopIO(closer io.Closer) {
	go func() {
		select {
		case <-i.ctx.Done():
			i.err = ErrIOTerminated
			closer.Close()

		case <-i.done:
		}
	}()
}

// Reader implements chan'ing for an io.Reader object.
type Reader struct {
	*IO
	rd      io.Reader
	bufsize int
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
		IO:      newIO(ctx),
		output:  make(chan []byte),
	}

	r.readLoop()

	closer, ok := r.rd.(io.Closer)
	if ok {
		r.stopIO(closer)
	}

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

			nb_read, err := r.rd.Read(buf)
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
			case r.output <- buf[:nb_read]:
			case <-r.ctx.Done():
				return
			}
		}
	}()
}

type Writer struct {
	*IO
	wr    io.Writer
	input chan []byte
}

func NewWriter(wr io.Writer) *Writer {
	return NewWriterContext(wr, context.Background())
}

func NewWriterContext(wr io.Writer, ctx context.Context) *Writer {
	w := &Writer{
		wr:    wr,
		IO:    newIO(ctx),
		input: make(chan []byte),
	}

	w.writeLoop()

	closer, ok := w.wr.(io.Closer)
	if ok {
		w.stopIO(closer)
	}

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
