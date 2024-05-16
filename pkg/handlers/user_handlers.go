package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"risque-server/pkg/models"
	"risque-server/pkg/utils"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        log.Printf("Error decoding user: %v\n", err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    conn, err := utils.GetDBConnection()
    if err != nil {
        log.Println(err)
        http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
        return
    }

    sql := `INSERT INTO users (name, email) VALUES ($1, $2)`
    _, err = conn.Exec(r.Context(), sql, user.Name, user.Email)
    if err != nil {
        log.Printf("Error inserting user: %v\n", err)
        http.Error(w, "Unable to insert user", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(user) // Ensure you send back the inserted user data
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
    conn, err := utils.GetDBConnection()
    if err != nil {
        log.Printf("Error getting DB connection: %v\n", err)
        http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
        return
    }
    defer conn.Close()

    rows, err := conn.Query(r.Context(), "SELECT id, name, email FROM users")
    if err != nil {
        log.Printf("Error querying users: %v\n", err)
        http.Error(w, "Unable to query users", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var users []models.User
    for rows.Next() {
        var user models.User
        if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
            log.Printf("Error scanning user: %v\n", err)
            http.Error(w, "Error scanning row", http.StatusInternalServerError)
            return
        }
        users = append(users, user)
    }

    w.Header().Set("Content-Type", "application/json")
    
    if len(users) == 0 {
        w.Write([]byte("[]"))
        return
    }

    if err := json.NewEncoder(w).Encode(users); err != nil {
        log.Printf("Error encoding users to JSON: %v\n", err)
        http.Error(w, "Error encoding users", http.StatusInternalServerError)
        return
    }
}

