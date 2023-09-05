package repository

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/rhine-tech/scene/lens/infrastructure/ingestion"
)

type openobserve[T any] struct {
	org    string
	stream string
	client *resty.Client
}

func NewOpenObserveIngestor[T any](user, password string, baseUrl, org string) ingestion.Ingestor[T] {
	client := resty.New()
	client.SetBaseURL(baseUrl).SetBasicAuth(user, password)

	return &openobserve[T]{
		org:    org,
		stream: "default",
		client: client,
	}
}

func NewOpenObserveCommonIngestor(user, password string, baseUrl, org string) ingestion.CommonIngestor {
	return NewOpenObserveIngestor[any](user, password, baseUrl, org)
}

func (o *openobserve[T]) Ingest(msg ...T) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return ingestion.ErrFailedToParseData.WithDetail(err)
	}
	api := fmt.Sprintf("/api/%s/%s/_json", o.org, o.stream)
	_, err = o.client.R().SetHeader("content-type", "application/json").SetBody(data).Post(api)
	if err != nil {
		return ingestion.ErrFailedToIngestData.WithDetail(err)
	}
	return nil
}
func (o *openobserve[T]) UsePipe(pipe string) ingestion.Ingestor[T] {
	return &openobserve[T]{
		org:    o.org,
		stream: pipe,
		client: o.client,
	}
}
