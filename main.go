package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Initialize your data store
	initializeStore()

	// Initialize flags
	InitFlags()

	// Set up your HTTP routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/tasks", authenticateMiddleware(handleTasks))
	http.HandleFunc("/task/", authenticateMiddleware(handleTask))
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/admin", authenticateMiddleware(handleAdmin))
	http.HandleFunc("/system", handleSystem)
	http.HandleFunc("/check-flag", handleCheckFlag)

	// Start the server
	fmt.Println("Starting server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	welcomeFlag := GetFlag(WelcomeFlag)
	encodedFlag := base64.StdEncoding.EncodeToString([]byte(welcomeFlag))

	// Check if the user has already solved the first challenge
	alreadySolved := checkFirstFlag(r)

	var onloadScript, flagChallengeStyle, mainContentStyle string
	if alreadySolved {
		onloadScript = "window.onload = showMainContent;"
		flagChallengeStyle = "display: none;"
		mainContentStyle = "display: block;"
	} else {
		onloadScript = ""
		flagChallengeStyle = "display: block;"
		mainContentStyle = "display: none;"
	}

	htmlContent := fmt.Sprintf(`
    <html>
        <head>
            <title>The TaskForge Trials</title>
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
                ul { list-style-type: none; padding: 0; }
                li { margin-bottom: 10px; padding: 5px; background-color: #f8f9fa; border-radius: 3px; }
                form { margin-top: 20px; }
                input[type="text"], 
                input[type="password"] { 
                    width: 100%%; 
                    padding: 8px; 
                    margin-top: 10px;
                    border: 1px solid #ddd;
                    border-radius: 4px;
                }
                input[type="submit"], 
                button { 
                    padding: 10px 15px; 
                    background-color: #007bff; 
                    color: white; 
                    border: none; 
                    cursor: pointer; 
                    margin-top: 10px;
                    border-radius: 4px;
                }
                input[type="submit"]:hover, 
                button:hover {
                    background-color: #0056b3;
                }
                a { color: #007bff; text-decoration: none; }
                a:hover { text-decoration: underline; }
                #mainContent { display: none; }
                #flagChallenge {
                    text-align: center;
                    padding: 20px;
                    background-color: #f0f0f0;
                    border-radius: 5px;
                }
                #firstFlagInput {
                    width: 70%%;
                    display: inline-block;
                }
            </style>
            <script>
                function submitFlag() {
                    const flagInput = document.getElementById('firstFlagInput');
                    const flag = flagInput.value;
                    
                    console.log('Submitting flag:', flag);

                    fetch('/check-flag', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({ flag: flag })
                    })
                    .then(response => {
                        console.log('Flag check response status:', response.status);
                        return response.json();
                    })
                    .then(data => {
                        console.log('Flag check response:', data);
                        if (data.success) {
                            document.getElementById('flagChallenge').style.display = 'none';
                            document.getElementById('mainContent').style.display = 'block';
                            alert('Correct flag! You can now register and log in.');
                        } else {
                            alert('Incorrect flag. Try again!');
                        }
                    })
                    .catch(error => {
                        console.error('Error:', error);
                        alert('An error occurred. Please try again.');
                    });
                }

                function handleSubmit(event, formId) {
                    event.preventDefault();
                    const form = document.getElementById(formId);
                    const formData = new FormData(form);
                    
                    console.log('Submitting form:', formId);
                    for (let [key, value] of formData.entries()) {
                        console.log(key, value);
                    }

                    fetch(form.action, {
                        method: form.method,
                        body: formData
                    })
                    .then(response => {
                        console.log('Form submission response status:', response.status);
                        return response.json();
                    })
                    .then(data => {
                        console.log('Server response:', data);
                        if (data.success) {
                            alert(data.message);
                            if (formId === 'loginForm') {
                                window.location.href = '/tasks';
                            }
                        } else {
                            alert('Error: ' + (data.message || 'Unknown error occurred'));
                        }
                    })
                    .catch(error => {
                        console.error('Error:', error);
                        alert('An error occurred. Please try again.');
                    });
                }

                function showMainContent() {
                    document.getElementById('flagChallenge').style.display = 'none';
                    document.getElementById('mainContent').style.display = 'block';
                }

                %s
            </script>
        </head>
        <body>
            <h1>Welcome to the TaskForge Trials</h1>
            <p>Your journey begins here. Can you uncover all the secrets?</p>
            
            <div id="flagChallenge" class="challenge" style="%s">
                <input type="text" id="firstFlagInput" placeholder=". . . ">
                <button onclick="submitFlag()">Submit</button>
            </div>

            <div id="mainContent" style="%s">
                <div class="challenge">
                    <h2>Challenge 1: Register an Account</h2>
                    <form id="registerForm" action="/register" method="post" onsubmit="handleSubmit(event, 'registerForm')">
                        <input type="text" name="username" placeholder="Username" required>
                        <input type="password" name="password" placeholder="Password" required>
                        <input type="submit" value="Register">
                    </form>
                </div>

                <div class="challenge">
                    <h2>Challenge 2: Log In</h2>
                    <form id="loginForm" action="/login" method="post" onsubmit="handleSubmit(event, 'loginForm')">
                        <input type="text" name="username" placeholder="Username" required>
                        <input type="password" name="password" placeholder="Password" required>
                        <input type="submit" value="Log In">
                    </form>
                </div>

                <div class="challenge">
                    <h2>Challenge 3: Create a Task</h2>
                    <form id="taskForm" action="/tasks" method="post" onsubmit="handleSubmit(event, 'taskForm')">
                        <input type="text" name="title" placeholder="Task Title" required>
                        <input type="submit" value="Create Task">
                    </form>
                </div>

                <h2>How to Play</h2>
                <p>1. Register an account and log in.</p>
                <p>2. Try creating and viewing tasks.</p>
                <p>3. Look for hidden clues in the page source and network requests.</p>
                <p>4. As you progress, you may need to use more advanced techniques to uncover deeper secrets!</p>
            </div>

            <!-- Congratulations, adventurer! You've discovered: %s -->
        </body>
    </html>
    `,
		onloadScript,
		flagChallengeStyle,
		mainContentStyle,
		encodedFlag)

	fmt.Fprint(w, htmlContent)
}

func authenticateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		sessionToken := cookie.Value
		userSession := getSession(sessionToken)
		if userSession == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// You can add the username to the request context if needed
		ctx := context.WithValue(r.Context(), "username", userSession.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
