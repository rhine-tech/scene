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
}

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
	remaining := s.size - s.pos
	toRead := int64(len(p))
	if toRead > remaining {
		toRead = remaining
	}
	data, err := s.srv.Load(s.fileId, s.pos, toRead)
	if err != nil {
		return 0, err
	}
	n := copy(p, data)
	s.pos += int64(n)
	if int64(n) < toRead {
		return n, io.EOF
	}
	return n, nil
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
	return s.pos, nil
}
