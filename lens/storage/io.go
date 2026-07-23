package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
)

var errContentReaderClosed = errors.New("storage content reader is closed")

type contentReader struct {
	ctx        context.Context
	srv        IStorageService
	storageKey StorageKey
	pos        int64
	size       int64
	current    io.ReadCloser
	closed     bool
}

// OpenContent opens storage content as a request-scoped seekable stream.
// Sequential reads share one provider reader. Seeking closes that reader and
// opens a new ranged reader lazily on the next Read call.
func OpenContent(ctx context.Context, srv IStorageService, storageKey StorageKey) (io.ReadSeekCloser, FileMeta, error) {
	meta, err := srv.Meta(ctx, storageKey)
	if err != nil {
		return nil, meta, err
	}
	if meta.ContentLength < 0 {
		return nil, meta, fmt.Errorf("invalid content length %d", meta.ContentLength)
	}
	return &contentReader{
		ctx:        ctx,
		srv:        srv,
		storageKey: storageKey,
		size:       meta.ContentLength,
	}, meta, nil
}

func (r *contentReader) Read(p []byte) (int, error) {
	if r.closed {
		return 0, errContentReaderClosed
	}
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	if len(p) == 0 {
		return 0, nil
	}
	if r.pos >= r.size {
		return 0, io.EOF
	}

	remaining := r.size - r.pos
	if int64(len(p)) > remaining {
		p = p[:int(remaining)]
	}
	if r.current == nil {
		reader, err := r.srv.Load(r.ctx, r.storageKey, r.pos, remaining)
		if err != nil {
			return 0, err
		}
		if reader == nil {
			return 0, errors.New("storage service returned a nil content reader")
		}
		r.current = reader
	}

	n, err := r.current.Read(p)
	r.pos += int64(n)
	if errors.Is(err, io.EOF) && r.pos < r.size {
		return n, io.ErrUnexpectedEOF
	}
	return n, err
}

func (r *contentReader) Seek(offset int64, whence int) (int64, error) {
	if r.closed {
		return r.pos, errContentReaderClosed
	}

	var newPos int64
	switch whence {
	case io.SeekStart:
		if offset < 0 || offset > r.size {
			return r.pos, errors.New("invalid seek position")
		}
		newPos = offset
	case io.SeekCurrent:
		if offset < -r.pos || offset > r.size-r.pos {
			return r.pos, errors.New("invalid seek position")
		}
		newPos = r.pos + offset
	case io.SeekEnd:
		if offset < -r.size || offset > 0 {
			return r.pos, errors.New("invalid seek position")
		}
		newPos = r.size + offset
	default:
		return r.pos, errors.New("invalid seek whence")
	}

	if newPos == r.pos {
		return r.pos, nil
	}
	if err := r.closeCurrent(); err != nil {
		return r.pos, err
	}
	r.pos = newPos
	return r.pos, nil
}

func (r *contentReader) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	return r.closeCurrent()
}

func (r *contentReader) closeCurrent() error {
	if r.current == nil {
		return nil
	}
	reader := r.current
	r.current = nil
	return reader.Close()
}
