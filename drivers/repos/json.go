package repos

import (
	"encoding/json"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/model"
	"os"
)

// Deprecated: use JsonDatasourceRepo one instead
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

func (j *CommonJsonRepo[T]) RepoImplName() scene.ImplName {
	return scene.NewRepoImplName("json", "JsonRepository", "json")
}

func (j *CommonJsonRepo[T]) Status() error {
	return j.Err
}

// Deprecated: use JsonDatasourceRepo one instead
func NewJsonRepository[T any](cfg model.FileConfig) *CommonJsonRepo[T] {
	rp := &CommonJsonRepo[T]{
		Cfg: cfg,
	}
	return rp
}

type JsonDatasourceRepo[T any] struct {
	datasource datasource.JsonDataSource
	Data       T
}

func UseJsonDatasourceRepo[T any](datasource datasource.JsonDataSource) *JsonDatasourceRepo[T] {
	rp := &JsonDatasourceRepo[T]{datasource: datasource, Data: *new(T)}
	return rp
}

func (j *JsonDatasourceRepo[T]) LoadData() (T, error) {
	var data T
	content, err := j.datasource.Load()
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(content, &data)
	if err != nil {
		return data, err
	}
	j.Data = data
	return data, err
}
