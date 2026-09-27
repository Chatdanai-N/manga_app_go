package models

type MangaSystemParameter struct {
	ParameterId uint64 `gorm:"primaryKey;column:parameter_id" json:"parameterId"`
	GroupCode   string `gorm:"column:group_code" json:"groupCode"`
	Code        string `gorm:"column:code" json:"code"`
	Value       string `gorm:"column:value" json:"value"`
	OrderNumber uint64 `gorm:"column:order_number" json:"orderNumber"`
}

func (MangaSystemParameter) TableName() string {
	return "manga_system_parameter"
}
