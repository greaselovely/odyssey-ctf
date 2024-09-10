package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// handleTasks manages the tasks for a user.
// 1. It allows users to get their tasks and create new ones.
// 2. The GET method retrieves all tasks for the authenticated user.
// 3. The POST method allows creation of new tasks.
func handleTasks(w http.ResponseWriter, r *http.Request) {
	if !checkFirstFlag(r) {
		sendJSONResponse(w, false, "You must solve the first challenge to access tasks", http.StatusForbidden)
		return
	}
	// Get the username from the session
	username, err := getUsernameFromSession(r)
	if err != nil {
		sendJSONResponse(w, false, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Retrieve tasks for the user
		tasks := GetTasksByUsername(username)
		displayTasks(w, r, tasks)

	case http.MethodPost:
		// Parse the form data
		if err := r.ParseForm(); err != nil {
			sendJSONResponse(w, false, "Error parsing form: "+err.Error(), http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		if title == "" {
			sendJSONResponse(w, false, "Task title is required", http.StatusBadRequest)
			return
		}

		// Create a new task
		newTask := &Task{
			Title:     title,
			Completed: false,
			OwnerID:   GetUserByUsername(username).ID,
		}
		AddTask(newTask)

		// Check for task injection flag
		if strings.Contains(title, "'") {
			log.Printf("Task Injection Flag triggered by user: %s", username)
			// In a real scenario, you might want to set a flag or trigger an event here
		}

		// Redirect to the tasks page
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)

	default:
		sendJSONResponse(w, false, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func displayTasks(w http.ResponseWriter, r *http.Request, tasks []Task) {
	w.Header().Set("Content-Type", "text/html")
	html := `
    <html>
        <head>
            <title>Your Tasks</title>
            <style>
                body { 
                    font-family: Arial, sans-serif; 
                    line-height: 1.6; 
                    padding: 20px; 
                    max-width: 800px; 
                    margin: 0 auto; 
                }
                h1, h2 { color: #2c3e50; }
                .challenge { 
                    margin-bottom: 20px; 
                    padding: 10px; 
                    border: 1px solid #ddd; 
                    border-radius: 5px;
                }
                ul { 
                    list-style-type: none; 
                    padding: 0; 
                }
                li { 
                    margin-bottom: 10px; 
                    padding: 10px; 
                    background-color: #f8f9fa; 
                    border-radius: 3px; 
                    border: 1px solid #e9ecef;
                }
                form { margin-top: 20px; }
                input[type="text"] { 
                    width: 100%; 
                    padding: 8px; 
                    margin-top: 10px;
                    border: 1px solid #ddd;
                    border-radius: 4px;
                }
                input[type="submit"] { 
                    padding: 10px 15px; 
                    background-color: #007bff; 
                    color: white; 
                    border: none; 
                    cursor: pointer; 
                    margin-top: 10px;
                    border-radius: 4px;
                }
                input[type="submit"]:hover {
                    background-color: #0056b3;
                }
                a { 
                    color: #007bff; 
                    text-decoration: none; 
                    display: inline-block;
                    margin-bottom: 10px;
                }
                a:hover { text-decoration: underline; }
            </style>
        </head>
        <body>
            <h1>Your Tasks</h1>
            <a href="/">Home</a>
            <div class="challenge">
                <h2>Task List</h2>
                <ul>
    `
	if len(tasks) == 0 {
		html += "<li>No tasks found. Create a new task below!</li>"
	} else {
		for _, task := range tasks {
			html += fmt.Sprintf("<li>ID: %d - %s</li>", task.ID, task.Title)
		}
	}
	html += `
                </ul>
            </div>
            <div class="challenge">
                <h2>Add New Task</h2>
                <form action="/tasks" method="post">
                    <input type="text" name="title" placeholder="Task Title" required>
                    <input type="submit" value="Create Task">
                </form>
            </div>
        </body>
    </html>
    `
	fmt.Fprint(w, html)
}

// handleTask manages operations on individual tasks.
// CTF Vulnerability: This function doesn't verify task ownership.
// An attacker could potentially access or modify any task if they know its ID.
func handleTask(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement single task handling
	fmt.Fprintf(w, "Single task endpoint")
}

// handleLogin processes user login attempts.
// CTF Vulnerability: This function compares passwords in plain text.
// In a real application, passwords should be hashed for security.
func handleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("Login request received")
	log.Printf("Request Content-Type: %s", r.Header.Get("Content-Type"))
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		sendJSONResponse(w, false, "Error reading request", http.StatusInternalServerError)
		return
	}
	log.Printf("Request Body: %s", string(body))
	r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset the body for later use

	if !checkFirstFlag(r) {
		log.Println("First flag check failed")
		sendJSONResponse(w, false, "You must solve the first challenge to log in", http.StatusForbidden)
		return
	}

	log.Println("First flag check passed")

	var username, password string

	// Check the Content-Type of the request
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		// Parse JSON data
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			sendJSONResponse(w, false, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		username = creds.Username
		password = creds.Password
	} else {
		// Parse form data
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			log.Printf("Error parsing multipart form: %v", err)
			if err := r.ParseForm(); err != nil {
				log.Printf("Error parsing form: %v", err)
				sendJSONResponse(w, false, "Error parsing form: "+err.Error(), http.StatusBadRequest)
				return
			}
		}
		username = r.FormValue("username")
		password = r.FormValue("password")
		log.Printf("Parsed username: %s, password length: %d", username, len(password))
	}

	// Validate input
	if username == "" || password == "" {
		sendJSONResponse(w, false, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Authenticate user
	user := GetUserByUsername(username)
	log.Printf("Login attempt for user: %s", username)
	if user == nil || user.Password != password { // Note: In a real app, you should use proper password hashing
		log.Printf("Login failed for user: %s", username)
		sendJSONResponse(w, false, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("Login successful for user: %s", username)

	// Get the existing session token from the cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		// If there's no session, create a new one
		sessionToken := createSession(username)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: time.Now().Add(24 * time.Hour),
		})
	} else {
		// Update the existing session with the username
		updateSession(cookie.Value, username)
	}

	sendJSONResponse(w, true, fmt.Sprintf("Welcome, %s!", username), http.StatusOK)
}

// handleRegister processes new user registrations.
// It ensures new users are always created with a 'user' role for security.
func handleRegister(w http.ResponseWriter, r *http.Request) {
	if !checkFirstFlag(r) {
		sendJSONResponse(w, false, "You must solve the first challenge to register", http.StatusForbidden)
		return
	}
	log.Printf("Request Method: %s", r.Method)
	log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))

	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		sendJSONResponse(w, false, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Log all form values
	log.Printf("Form values: %v", r.MultipartForm.Value)

	// Get username and password from form data
	username := r.FormValue("username")
	password := r.FormValue("password")

	log.Printf("Username: %s, Password length: %d", username, len(password))

	// Validate input
	if username == "" || password == "" {
		log.Printf("Username or password is empty")
		sendJSONResponse(w, false, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	if GetUserByUsername(username) != nil {
		log.Printf("Username already exists: %s", username)
		sendJSONResponse(w, false, "Username already exists", http.StatusConflict)
		return
	}

	// Create new user
	newUser := User{
		Username: username,
		Password: password, // Note: In a real application, you should hash this password
		Role:     "user",
	}
	AddUser(newUser)

	log.Printf("User registered successfully: %s", username)
	// Respond with success JSON
	sendJSONResponse(w, true, fmt.Sprintf("User %s registered successfully", username), http.StatusCreated)
}

// Helper function to send JSON responses
func sendJSONResponse(w http.ResponseWriter, success bool, message string, statusCode int, data ...map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"success": success,
		"message": message,
	}

	if len(data) > 0 {
		for key, value := range data[0] {
			response[key] = value
		}
	}

	json.NewEncoder(w).Encode(response)
}

// handleAdmin provides an admin-only endpoint to view all tasks.
// CTF Challenge: Participants might need to find ways to escalate their privileges to access this.
func handleAdmin(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement admin functionality
	fmt.Fprintf(w, "Admin endpoint")
}

// handleSystem provides a vulnerable endpoint that allows command injection.
// CTF Vulnerability: This function has no authentication check and allows for command injection.
// This is a severe vulnerability that should never exist in a real application.
func handleSystem(w http.ResponseWriter, r *http.Request) {
	// CTF Vulnerability: No authentication check
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		hint := "Hint: Data must be submitted as JSON. Use -H \"Content-Type: application/json\" in your curl command."
		http.Error(w, hint, http.StatusBadRequest)
		return
	}

	var cmd struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		hint := "Hint: Invalid JSON format. Ensure your JSON is correctly formatted and includes a 'command' field."
		http.Error(w, hint, http.StatusBadRequest)
		return
	}

	var response string
	// NEVER do this in a real application!
	if cmd.Command == "get_flag" {
		response = GetFlag(SystemCompromisedFlag)
	} else {
		response = fmt.Sprintf("Unknown command: %s. >> Use get_flag instead <<", cmd.Command)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"response": response})
}

func handleCheckFlag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var flagData struct {
		Flag string `json:"flag"`
	}
	if err := json.NewDecoder(r.Body).Decode(&flagData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	correctFlag := GetFlag(WelcomeFlag)
	if flagData.Flag == correctFlag {
		// Create a session to store that the user has solved the first challenge
		sessionToken := createSession("") // Empty string as we don't have a username yet
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: time.Now().Add(24 * time.Hour),
		})

		// Mark the session as having solved the first challenge
		markSessionSolvedFirstChallenge(sessionToken)

		sendJSONResponse(w, true, "Correct flag!", http.StatusOK)
	} else {
		sendJSONResponse(w, false, "Incorrect flag", http.StatusUnauthorized)
	}
}
