package models

// type Student struct {
// 	ID    string `json:"id" binding:"required"`
// 	Name  string `json:"name" binding:"required"`
// 	Tel   string `json:"tel"`
// 	Study string `json:"study"`
// }

// Student 模型 (添加 GORM 标签)
type Student struct {
	ID    string `json:"id" gorm:"primaryKey;column:id"`   // 指定为主键和列名
	Name  string `json:"name" gorm:"column:name;not null"` // not null 约束
	Tel   string `json:"tel" gorm:"column:tel"`            // 可为空
	Study string `json:"study" gorm:"column:study"`        // 可为空
}
