package main

import (
	"sync"
)

var (
	users      []User
	tasks      []Task
	userMutex  sync.RWMutex
	taskMutex  sync.RWMutex
	nextUserID = 1
	nextTaskID = 1
)

func initializeStore() {
	// Initialize with a default admin user
	users = append(users, User{
		ID:       getNextUserID(),
		Username: "admin",
		Password: "admin123", // In a real app, this should be hashed
		Role:     "admin",
	})

	// Initialize with a sample task
	tasks = append(tasks, Task{
		ID:        getNextTaskID(),
		Title:     "Find the first flag",
		Completed: false,
		OwnerID:   1,
	})

	// Hidden task (potential CTF element)
	tasks = append(tasks, Task{
		ID:        getNextTaskID(),
		Title:     "CTF{hidden_task_found}",
		Completed: false,
		OwnerID:   0, // No owner, hidden task
	})
}

func getNextUserID() int {
	id := nextUserID
	nextUserID++
	return id
}

func getNextTaskID() int {
	id := nextTaskID
	nextTaskID++
	return id
}

// AddUser adds a new user to the store
func AddUser(user User) {
	userMutex.Lock()
	defer userMutex.Unlock()
	user.ID = getNextUserID()
	users = append(users, user)
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(username string) *User {
	userMutex.RLock()
	defer userMutex.RUnlock()
	for _, u := range users {
		if u.Username == username {
			return &u
		}
	}
	return nil
}

// AddTask adds a new task to the store
func AddTask(task *Task) {
	taskMutex.Lock()
	defer taskMutex.Unlock()
	task.ID = getNextTaskID()
	tasks = append(tasks, *task)
}

// GetTasksByOwner retrieves all tasks for a given owner ID
func GetTasksByOwner(ownerID int) []Task {
	taskMutex.RLock()
	defer taskMutex.RUnlock()
	var ownerTasks []Task
	for _, t := range tasks {
		if t.OwnerID == ownerID {
			ownerTasks = append(ownerTasks, t)
		}
	}
	return ownerTasks
}

// GetAllTasks retrieves all tasks (admin function)
func GetAllTasks() []Task {
	taskMutex.RLock()
	defer taskMutex.RUnlock()
	return tasks
}

// GetTasksByUsername retrieves all tasks for a given username
func GetTasksByUsername(username string) []Task {
	user := GetUserByUsername(username)
	if user == nil {
		return nil
	}
	return GetTasksByOwner(user.ID)
}
