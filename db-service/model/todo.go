package model

type TodoItem struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
	Id    string `json:"id"`
}
