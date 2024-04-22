package ingestion

import "github.com/rhine-tech/scene/errcode"

var _eg = errcode.NewErrorGroup(3, "ingestion")

var (
	ErrFailedToParseData  = _eg.CreateError(1, "failed to parse data")
	ErrFailedToIngestData = _eg.CreateError(2, "failed to ingest data")
)

type Ingestor[T any] interface {
	Ingest(msg ...T) error
	UsePipe(pipe string) Ingestor[T]
}

type CommonIngestor Ingestor[any]
