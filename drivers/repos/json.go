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

func (j *CommonJsonRepo[T]) Setup() error {
	content, err := os.ReadFile(j.Cfg.Path)
	var data T
	if err != nil {
		j.Err = err
		return err
	}
	j.Err = json.Unmarshal(content, &data)
	j.Data = data
	return j.Err
}

func (j *CommonJsonRepo[T]) Dispose() error {
	// store data to file
	content, err := json.MarshalIndent(j.Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(j.Cfg.Path, content, 0644)
}

func (j *CommonJsonRepo[T]) RepoImplName() string {
	return "json"
}

func (j *CommonJsonRepo[T]) Status() error {
	return j.Err
}

func NewJsonRepository[T any](cfg model.FileConfig) *CommonJsonRepo[T] {
	rp := &CommonJsonRepo[T]{
		Cfg: cfg,
	}
	return rp
}
