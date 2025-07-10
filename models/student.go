package models

type Student struct {
	ID    string `json:"id" binding:"required"`
	Name  string `json:"name" binding:"required"`
	Tel   string `json:"tel"`
	Study string `json:"study"`
}
