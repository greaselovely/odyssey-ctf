package main

// User represents a user in the system
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"` // The "-" means this field won't be included in JSON output
	Role     string `json:"role"`
}

// Task represents a task in the system
type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	OwnerID   int    `json:"owner_id"`
}
