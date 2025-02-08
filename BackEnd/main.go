package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/handlers"
	_ "github.com/mattn/go-sqlite3" // Importing for side effects
)

type UserStats struct {
	TotalSolved    int    `json:"totalSolved"`
	EasySolved     int    `json:"easySolved"`
	MediumSolved   int    `json:"mediumSolved"`
	HardSolved     int    `json:"hardSolved"`
	Submissions    int    `json:"submissions"`
	AcceptanceRate string `json:"acceptanceRate"`
}

type User struct {
	Email             string    `json:"email"`
	Password          string    `json:"password"`
	Name              string    `json:"name"`
	Username          string    `json:"username"`
	Location          string    `json:"location"`
	GitHub            string    `json:"github"`
	Twitter           string    `json:"twitter"`
	JoinDate          string    `json:"joinDate"`
	Stats             UserStats `json:"stats"`
	Bio               string    `json:"bio"`
	ProfileCompletion int       `json:"profileCompletion"`
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./users.db")
	if err != nil {
		panic(err)
	}
	// Create users table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		email TEXT PRIMARY KEY,
		password TEXT,
		name TEXT,
		username TEXT,
		location TEXT,
		github TEXT,
		twitter TEXT,
		joinDate TEXT,
		bio TEXT,
		profileCompletion INTEGER,
		totalSolved INTEGER,
		easySolved INTEGER,
		mediumSolved INTEGER,
		hardSolved INTEGER,
		submissions INTEGER,
		acceptanceRate TEXT
	)`)
	if err != nil {
		panic(err)
	}
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Initialize user stats to zero
		user.Stats = UserStats{
			TotalSolved:    0,
			EasySolved:     0,
			MediumSolved:   0,
			HardSolved:     0,
			Submissions:    0,
			AcceptanceRate: "0%",
		}
		// Store user data in the database
		_, err = db.Exec(`INSERT INTO users (email, password, name, username, location, github, twitter, joinDate, bio, profileCompletion, totalSolved, easySolved, mediumSolved, hardSolved, submissions, acceptanceRate) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			user.Email, user.Password, user.Name, user.Username, user.Location, user.GitHub, user.Twitter, user.JoinDate, user.Bio, user.ProfileCompletion,
			user.Stats.TotalSolved, user.Stats.EasySolved, user.Stats.MediumSolved, user.Stats.HardSolved, user.Stats.Submissions, user.Stats.AcceptanceRate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var storedPassword string
		err = db.QueryRow(`SELECT password FROM users WHERE email = ?`, user.Email).Scan(&storedPassword)
		if err != nil || storedPassword != user.Password {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		userData := getUserData(user.Email)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(userData)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func getUserData(email string) User {
	var user User
	row := db.QueryRow(`SELECT email, name, username, location, github, twitter, joinDate, bio, profileCompletion, totalSolved, easySolved, mediumSolved, hardSolved, submissions, acceptanceRate 
		FROM users WHERE email = ?`, email)
	err := row.Scan(&user.Email, &user.Name, &user.Username, &user.Location, &user.GitHub, &user.Twitter, &user.JoinDate, &user.Bio, &user.ProfileCompletion,
		&user.Stats.TotalSolved, &user.Stats.EasySolved, &user.Stats.MediumSolved, &user.Stats.HardSolved, &user.Stats.Submissions, &user.Stats.AcceptanceRate)
	if err != nil {
		return User{} // Return an empty user if not found
	}
	return user
}

func forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var request struct {
			Email string `json:"email"`
		}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var storedPassword string
		err = db.QueryRow(`SELECT password FROM users WHERE email = ?`, request.Email).Scan(&storedPassword)
		if err != nil {
			http.Error(w, "Email not found", http.StatusNotFound)
			// Generate a temporary password or token instead of sending the actual password
			tempPassword := "temporaryPassword123" // This should be securely generated
			// Simulate sending an email (you need to implement this function)
			err = simulateSendEmail(request.Email, tempPassword)
			if err != nil {
				http.Error(w, "Failed to send email", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode("Password sent to your email")
			return
		}
		http.Error(w, "Email not found", http.StatusNotFound)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// Simulate sending an email (placeholder function)
func simulateSendEmail(email, tempPassword string) error {
	// Here you would implement the actual email sending logic
	// For example, using an SMTP server or a third-party email service
	// This is a placeholder implementation
	return nil // Return nil to indicate success for now
}

func main() {
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/forgot-password", forgotPasswordHandler)

	// CORS configuration
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:5173"}),  // Allow your frontend origin
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}), // Allow specific methods
		handlers.AllowedHeaders([]string{"Content-Type"}),           // Allow specific headers
	)(http.DefaultServeMux)

	http.ListenAndServe(":8080", corsHandler)
}
