package model

// PaginationResult is a *interface* struct
// provide interface for all common pagination response
type PaginationResult[T any] struct {
	Total   int `json:"total" bson:"total"`
	Results []T `json:"results" bson:"results"`
	Offset  int `json:"offset" bson:"offset"`
}
