package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	model "achievesomethingbro/appmodel"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	// "golang.org/x/crypto/bcrypt" // Use this for production password comparison
)

func handleLoginRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"message": "Only POST method is allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Basic Validation
	if req.UserName == "" || req.Password == "" {
		http.Error(w, `{"message": "Username and password are required"}`, http.StatusBadRequest)
		return
	}
	log.Printf("Login Request: %+v", req)
	// Determine if the username is an email or mobile number
	query := "SELECT id, password FROM users WHERE user_name = ? and user_type=?"

	var userID int
	var storedPassword string

	// Execute query
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("DB Connection Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	err = db.QueryRow(query, req.UserName, req.UserType).Scan(&userID, &storedPassword)
	if err == sql.ErrNoRows {
		http.Error(w, `{"message": "Invalid username or password"}`, http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Printf("DB Query Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	log.Printf("Stored Password: %s, Hashed Input Password: %s", storedPassword, req.Password)
	if err != nil {
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	if req.Password != storedPassword {
		http.Error(w, `{"message": "Invalid username or password"}`, http.StatusUnauthorized)
		return
	}

	// Authentication successful
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user_id": userID,
	})
	db.Close()
}
