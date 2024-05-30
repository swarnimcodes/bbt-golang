package user

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/swarnimcodes/bitespeed-backend-task/utils"
)

func GetAllCustomers(w http.ResponseWriter, r *http.Request) {
	dbLocation := os.Getenv("DATABASE_LOCATION")
	if dbLocation == "" {
		message := "No database location specified"
		statusCode := http.StatusInternalServerError
		utils.SendErrorResponse(w, message, statusCode)
		return
	}

	// Ensure database location directory exists
	if err := os.MkdirAll(filepath.Dir(dbLocation), 0755); err != nil {
		message := fmt.Sprintf("Failed to create directories for database creation: %v", err)
		statusCode := http.StatusInternalServerError
		utils.SendErrorResponse(w, message, statusCode)
		return
	}

	// Open or create the .db file
	// Establish connection with the database
	db, err := sql.Open("sqlite3", dbLocation)
	if err != nil {
		message := fmt.Sprintf("Failed to connect to the database: %v", err)
		statusCode := http.StatusInternalServerError
		utils.SendErrorResponse(w, message, statusCode)
		return
	}
	defer db.Close()

	// Create user table if it doesn't exist
	if err := createCustomerTable(db); err != nil {
		message := fmt.Sprintf("Failed to create user table in the database: %v", err)
		statusCode := http.StatusInternalServerError
		utils.SendErrorResponse(w, message, statusCode)
		return
	}

	query := `
		SELECT *
		FROM Contact	
	`

	rows, err := db.Query(query)
	if err != nil {
	}
	var customers []Customer

	for rows.Next() {
		var customer Customer
		err := rows.Scan(
			&customer.Id,
			&customer.PhoneNumber,
			&customer.Email,
			&customer.LinkedId,
			&customer.LinkPrecedence,
			&customer.CreatedAt,
			&customer.UpdatedAt,
			&customer.DeletedAt,
		)
		if err != nil {

		}
		customers = append(customers, customer)
	}

	utils.SendJSONMessageResponse(w, customers, http.StatusOK)
}
