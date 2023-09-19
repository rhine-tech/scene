package model

type PageParam struct {
	Page int `json:"page" bson:"page"` // Page is the page number, starts from 1
	Size int `json:"size" bson:"size"` // Size is the page size
}

// PaginationParam is a *interface* struct
// provide interface for all common pagination parameter
type PaginationParam struct {
	Offset int `json:"offset" bson:"offset"` // Page is the page number, starts from 0
	Limit  int `json:"limit" bson:"limit"`   // Size is the page size
}
