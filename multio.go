package chanio

import (
	"context"
	"io"
	"sync"
)

type MultiReader struct {
	readers []*Reader
	bufsize int
	ctx     context.Context
	output  chan []byte
}

func NewMultiReader(readers ...io.Reader) *MultiReader {
	return NewMultiReaderSizeContext(defaultBufSize, context.Background(), readers...)
}

func NewMultiReaderSize(bufSize int, readers ...io.Reader) *MultiReader {
	return NewMultiReaderSizeContext(bufSize, context.Background(), readers...)
}

func NewMultiReaderContext(ctx context.Context, readers ...io.Reader) *MultiReader {
	return NewMultiReaderSizeContext(defaultBufSize, ctx, readers...)
}

func NewMultiReaderSizeContext(bufsize int, ctx context.Context, readers ...io.Reader) *MultiReader {
	mr := &MultiReader{
		bufsize: bufsize,
		ctx:     ctx,
		output:  make(chan []byte),
	}

	mr.readers = make([]*Reader, len(readers))

	for i, reader := range readers {
		mr.readers[i] = NewReaderSizeContext(reader, bufsize, ctx)
	}

	mr.readLoop()

	return mr
}

func (mr *MultiReader) Read() <-chan []byte {
	return mr.output
}

func (mr *MultiReader) readLoop() {
	var wg sync.WaitGroup

	for _, reader := range mr.readers {
		wg.Add(1)
		go func(reader *Reader) {
			defer wg.Done()
			for {
				select {
				case buf, ok := <-reader.Read():
					if !ok {
						return
					}
					select {
					case mr.output <- buf:
					case <-reader.done:
						return
					}

				case <-reader.done:
					return
				}
			}
		}(reader)
	}

	go func() {
		wg.Wait()
		close(mr.output)
	}()
}

func (mr *MultiReader) GetErrors() []error {
	errors := make([]error, len(mr.readers))

	for i, reader := range mr.readers {
		errors[i] = reader.GetError()
	}

	return errors
}
