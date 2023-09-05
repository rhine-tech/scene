package repos

import (
	"encoding/json"
	"github.com/rhine-tech/scene/model"
	"os"
)

type CommonJsonRepo[T any] struct {
	Cfg  model.FileConfig
	Data T
	Err  error
}

func (j *CommonJsonRepo[T]) RepoImplName() string {
	return "json"
}

func (j *CommonJsonRepo[T]) Status() error {
	return j.Err
}

func (j *CommonJsonRepo[T]) Dispose() error {
	return nil
}

func NewJsonRepository[T any](cfg model.FileConfig) *CommonJsonRepo[T] {
	rp := &CommonJsonRepo[T]{
		Cfg: cfg,
	}
	content, err := os.ReadFile(rp.Cfg.Path)
	var data T
	if err != nil {
		rp.Err = err
		return rp
	}
	rp.Err = json.Unmarshal(content, &data)
	rp.Data = data
	return rp
}
