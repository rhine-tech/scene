package repository

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/model"
	"os"
)

type JsonRepo struct {
	cfg  model.FileConfig
	data []byte
	log  logger.ILogger `aperture:""`
}

func NewJsonDataSource(cfg model.FileConfig) datasource.JsonDataSource {
	return &JsonRepo{cfg: cfg}
}

func NewJsonDataSourceFromPath(path string) datasource.JsonDataSource {
	return NewJsonDataSource(model.FileConfig{Path: path})
}

func (j *JsonRepo) Dispose() error {
	err := os.WriteFile(j.cfg.Path, j.data, 0644)
	if err != nil {
		j.log.Warnf("fail to write data to %s", j.cfg.Path)
	} else {
		j.log.Infof("save data to %s succeed", j.cfg.Path)
	}
	return nil
}

func (j *JsonRepo) Setup() error {
	j.log = j.log.WithPrefix(j.DataSourceName().String())
	if _, err := os.Stat(j.cfg.Path); os.IsNotExist(err) {
		j.log.Errorf("fail to open %s", j.cfg.Path)
		return err
	}
	return nil
}

func (j *JsonRepo) DataSourceName() scene.ImplName {
	return scene.NewRepoImplNameNoVer("datasource", "json")
}

func (j *JsonRepo) Status() error {
	return nil
}

func (j *JsonRepo) Load() ([]byte, error) {
	if j.data != nil {
		return j.data, nil
	}
	j.log.Infof("loading data from %s", j.cfg.Path)
	d, err := os.ReadFile(j.cfg.Path)
	if err != nil {
		return nil, err
	}
	j.data = d
	return d, nil
}

func (j *JsonRepo) Save(data []byte) error {
	j.data = data
	return nil
}
