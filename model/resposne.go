package model

type PageResult[T any] struct {
	Results   []T `json:"results" bson:"results"`
	PageNum   int `json:"page_num" bson:"page_num"`
	PageTotal int `json:"page_total" bson:"page_total"`
}

// PaginationResult is a *interface* struct
// provide interface for all common pagination response
type PaginationResult[T any] struct {
	Total   int64 `json:"total" bson:"total"`
	Offset  int64 `json:"offset" bson:"offset"`
	Results []T   `json:"results" bson:"results"`
	Count   int64 `json:"count" bson:"count"`
}

type JsonResponse map[string]interface{}

func PaginationResultTransform[Src any, Dst any](
	dst *PaginationResult[Dst], src *PaginationResult[Src],
	f func(dst *Dst, src *Src)) {
	dst.Total = src.Total
	dst.Offset = src.Total
	dst.Count = src.Count
	dst.Results = make([]Dst, len(src.Results))
	for idx, val := range src.Results {
		f(&dst.Results[idx], &val)
	}
}
