package handlers

import "sync"

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// Global variables moved here (starting with Uppercase to be exported)
var (
	Todos  []*Todo
	NextID = 1
	Mu     sync.RWMutex
)
