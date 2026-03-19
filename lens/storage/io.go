package storage

import (
	"errors"
	"io"
)

type IoInterface interface {
	io.Reader
	io.Seeker
}

type ioImpl struct {
	srv    IStorageService
	fileId FileID
	pos    int64
	size   int64
	buf    []byte
	bufPos int64
}

const defaultReadAheadSize int64 = 1 << 20 // 1 MiB

func NewIoInterface(srv IStorageService, fileId FileID) (IoInterface, FileMeta, error) {
	meta, err := srv.Meta(fileId)
	if err != nil {
		return nil, meta, err
	}
	return &ioImpl{
		srv:    srv,
		fileId: fileId,
		pos:    0,
		size:   meta.ContentLength,
	}, meta, nil
}

func (s *ioImpl) Read(p []byte) (int, error) {
	if s.pos >= s.size {
		return 0, io.EOF
	}
	if !s.hasBufferedDataAt(s.pos) {
		remaining := s.size - s.pos
		toRead := int64(len(p))
		if toRead < defaultReadAheadSize {
			toRead = defaultReadAheadSize
		}
		if toRead > remaining {
			toRead = remaining
		}
		reader, err := s.srv.Load(s.fileId, s.pos, toRead)
		if err != nil {
			return 0, err
		}
		data, err := io.ReadAll(reader)
		if closeErr := reader.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		if err != nil {
			return 0, err
		}
		if len(data) == 0 {
			return 0, io.EOF
		}
		s.buf = data
		s.bufPos = s.pos
	}

	start := int(s.pos - s.bufPos)
	n := copy(p, s.buf[start:])
	s.pos += int64(n)
	if s.pos >= s.size && n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (s *ioImpl) hasBufferedDataAt(pos int64) bool {
	if len(s.buf) == 0 {
		return false
	}
	return pos >= s.bufPos && pos < s.bufPos+int64(len(s.buf))
}

func (s *ioImpl) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = s.pos + offset
	case io.SeekEnd:
		newPos = s.size + offset
	default:
		return 0, errors.New("invalid seek whence")
	}
	if newPos < 0 || newPos > s.size {
		return 0, errors.New("invalid seek position")
	}
	s.pos = newPos
	if !s.hasBufferedDataAt(newPos) {
		s.buf = nil
		s.bufPos = 0
	}
	return s.pos, nil
}
