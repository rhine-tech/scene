package model

type PageResult[T any] struct {
	Results   []T `json:"results" bson:"results"`
	PageNum   int `json:"page_num" bson:"page_num"`
	PageTotal int `json:"page_total" bson:"page_total"`
}

// PaginationResult is a *interface* struct
// provide interface for all common pagination response
type PaginationResult[T any] struct {
	Total   int `json:"total" bson:"total"`
	Results []T `json:"results" bson:"results"`
	Offset  int `json:"offset" bson:"offset"`
}
