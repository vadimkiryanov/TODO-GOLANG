package model

type TodoItem struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}
type TodoItemId struct {
	Id string `json:"id"`
}
