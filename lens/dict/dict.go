package dict

type DictItem struct {
	ID      string `json:"id"`      // 唯一标识符
	Type    string `json:"type"`    // 字典类型（例如 "user_status"）
	Key     string `json:"key"`     // 字典项 key
	Value   string `json:"value"`   // 字典项 value
	Enabled bool   `json:"enabled"` // 是否启用
}

type IDictService interface {
	ListByType(dictType string) ([]DictItem, error)
	//Query(dictType string, ) (model.PaginationResult[DictItem], error)
	Get(dictType string, key string) (DictItem, error)
	HasKey(dictType string, key string) (bool, error)
	Create(item DictItem) error
	Update(id string, item DictItem) error
	Enable(id string) error
	Disable(id string) error
	Delete(id string) error

	Import(dictType string, items []DictItem, override bool) error
}
