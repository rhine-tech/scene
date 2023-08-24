package repository

import "github.com/aynakeya/scene/lens/infrastructure/ingestion"

type dummy[T any] struct {
}

func NewDummyIngestor[T any]() ingestion.Ingestor[T] {
	return &dummy[T]{}
}

func NewDummyCommonIngestor() ingestion.CommonIngestor {
	return NewDummyIngestor[any]()
}

func (d *dummy[T]) Ingest(msg ...T) error {
	return nil
}

func (d *dummy[T]) UsePipe(pipe string) ingestion.Ingestor[T] {
	return d
}
