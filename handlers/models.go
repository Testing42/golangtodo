package handlers

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// Global variables moved here (starting with Uppercase to be exported)
var Todos []Todo
var NextID = 1
