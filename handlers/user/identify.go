package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/swarnimcodes/bitespeed-backend-task/utils"
)

// TODO: check how to store just two values "primary" and "secondary"
// datetime
func createCustomerTable(db *sql.DB) error {
	tableCreationQuery := `
		CREATE TABLE IF NOT EXISTS Contact (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		phoneNumber TEXT,
		email TEXT,
		linkedId INTEGER,
		linkPrecedence TEXT,
		createdAt TEXT NOT NULL, -- store ISO 8601 datetime
		updatedAt TEXT NOT NULL,
		deletedAt TEXT
		)
		`
	_, err := db.Exec(tableCreationQuery)
	return err
}
func getPrimaryIdForPartialMatch(db *sql.DB, phoneNumber, email string) (int, error) {
	relatedPrimaryId := -1
	query := `
		SELECT id
		FROM Contact
		WHERE linkPrecedence = 'primary'
		AND (phoneNumber = ? OR email = ? )
	`
	err := db.QueryRow(query, phoneNumber, email).Scan(&relatedPrimaryId)
	if err == sql.ErrNoRows {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}
	return relatedPrimaryId, nil
}

func doesUserExist(db *sql.DB, phoneNumber, email string) (bool, error) {
	count := 0
	query := `
		SELECT COUNT(*)
		FROM Contact
		WHERE phoneNumber = ? AND email = ?
	`
	err := db.QueryRow(query, phoneNumber, email).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 1 {
		err := errors.New("PROGRAM MALFUNCTIONED")
		return false, err
	}
	return count == 1, nil
}

func getCustomers(db *sql.DB, phoneNumber, email string) ([]Customer, error) {
	var primaryId int
	var linkedId sql.NullInt64

	// check if input matches a primary contact
	query := `
		SELECT id, linkedId
		FROM Contact
		WHERE (phoneNumber = ? OR email = ?)
		AND linkPrecedence = 'primary'
		LIMIT 1;
	`

	err := db.QueryRow(query, phoneNumber, email).Scan(&primaryId, &linkedId)
	if err == sql.ErrNoRows {
		// the input details did not match a primary contact
		// check if the details match a secondary contact
		// if match ==> find primary contact's id
		query := `
			SELECT linkedId
			FROM CONTACT
			WHERE (phoneNumber = ? OR email = ?)
			AND linkPrecedence = 'secondary'
			LIMIT 1
		`

		err := db.QueryRow(query, phoneNumber, email).Scan(&primaryId)
		if err == sql.ErrNoRows {
			// User insertion will be carried out by the caller
			return []Customer{}, nil
		} else if err != nil {
			return []Customer{}, err
		}
	}

	// get all contacts linked to the primary contact
	query = `
		SELECT id, phoneNumber, email, linkedId, linkPrecedence, createdAt, updatedAt, deletedAt
		FROM Contact
		WHERE id = ? OR linkedId = ?
		`
	rows, err := db.Query(query, primaryId, primaryId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
			return nil, err
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return customers, nil
}

func insertPrimaryNewCustomer(db *sql.DB, phoneNumber, email string) error {
	query := `
		INSERT INTO Contact (phoneNumber, email, linkedId, linkPrecedence, createdAt, updatedAt)
		VALUES (?, ?, NULL, "primary",  datetime('now'), datetime('now'))
	`
	_, err := db.Exec(query, phoneNumber, email)
	return err
}

func insertNewSecondaryCustomer(db *sql.DB, phoneNumber string, email string, relatedPrimaryId int) error {
	query := `
		INSERT INTO Contact (phoneNumber, email, linkedId, linkPrecedence, createdAt, updatedAt)
		VALUES (?, ?, ?, "secondary",  datetime('now'), datetime('now'))
	`
	_, err := db.Exec(query, phoneNumber, email, relatedPrimaryId)
	return err
}

func IdentifyCustomer(w http.ResponseWriter, r *http.Request) {
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

	// parse request body into User struct
	// var user User
	var customer Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		message := fmt.Sprintf("Invalid JSON request input payload: %v", err)
		statusCode := http.StatusBadRequest
		utils.SendErrorResponse(w, message, statusCode)
		return
	}

	// TODO: check if phone number or email exists in database.
	phoneNumber := customer.PhoneNumber
	email := customer.Email

	// FIRST STEP check if exact customer present in database
	// if yes ==> pass
	// if no ==> insert
	fmt.Println("bruh init")
	userExists, err := doesUserExist(db, *phoneNumber, *email)
	if err != nil {
		message := fmt.Sprintf("Error finding if user exists: %v", err)
		statusCode := http.StatusInternalServerError
		utils.SendErrorResponse(w, message, statusCode)
		return
	}
	relatedPrimaryId, err := getPrimaryIdForPartialMatch(db, *phoneNumber, *email)
	if err != nil {
		message := fmt.Sprintf("Error getting primary id: %v", err)
		statusCode := http.StatusInternalServerError
		utils.SendErrorResponse(w, message, statusCode)
		return
	}
	if !userExists && relatedPrimaryId > 0 {
		err := insertNewSecondaryCustomer(db, *phoneNumber, *email, relatedPrimaryId)
		if err != nil {
			message := fmt.Sprintf("Error inserting new secondary customer to database: %v", err)
			statusCode := http.StatusInternalServerError
			utils.SendErrorResponse(w, message, statusCode)
			return
		}
	} else if !userExists && relatedPrimaryId < 0 {
		// no such customer exists with either the phone number or the email
		// insert a new primary customer
		err := insertPrimaryNewCustomer(db, *phoneNumber, *email)
		if err != nil {
			message := fmt.Sprintf("Error inserting new customer to database: %v", err)
			statusCode := http.StatusInternalServerError
			utils.SendErrorResponse(w, message, statusCode)
			return
		}

		// retrieve the inserted user from the database
		var newCustomer Customer
		query := `
				SELECT id, phoneNumber, email, linkedId, linkPrecedence, createdAt, updatedAt, deletedAt
				FROM Contact
				WHERE phoneNumber = ? AND email = ? AND linkPrecedence = 'primary'
				LIMIT 1;
			`
		err = db.QueryRow(query, phoneNumber, email).Scan(
			&newCustomer.Id,
			&newCustomer.PhoneNumber,
			&newCustomer.Email,
			&newCustomer.LinkedId,
			&newCustomer.LinkPrecedence,
			&newCustomer.CreatedAt,
			&newCustomer.UpdatedAt,
			&newCustomer.DeletedAt,
		)
		if err != nil {
			// error
			message := fmt.Sprintf("Error fetching newly created customer: %v", err)
			statusCode := http.StatusInternalServerError
			utils.SendErrorResponse(w, message, statusCode)
			return
		}
		// here we have the newly inserted user and it will be primary
		// return this new user
		customerSummary := CustomerSummaryData{
			Contact: Contact{
				PrimaryContactId:    newCustomer.Id,
				Emails:              []string{*newCustomer.Email},
				PhoneNumbers:        []string{*newCustomer.PhoneNumber},
				SecondaryContactIds: []int{},
			},
		}

		// return the customer summary JSON
		utils.SendJSONMessageResponse(w, customerSummary, http.StatusOK)
		return
	} else {
		fmt.Println("bruh else")
	}

	// all customers that match either the phone number or email
	allMatchingCustomers, err := getCustomers(db, *phoneNumber, *email)
	fmt.Println("bruh all matching customers count: ", len(allMatchingCustomers))
	if err != nil {
		message := fmt.Sprintf("Could not fetch customers from database: %v", err)
		statusCode := http.StatusInternalServerError
		utils.SendErrorResponse(w, message, statusCode)
		return
	}

	// TODO: HANDLE CASE WHERE LEN == 2!!!!!
	if len(allMatchingCustomers) == 1 {
		currentCustomer := allMatchingCustomers[0]
		if allMatchingCustomers[0].LinkPrecedence == Primary {
			customerSummary := CustomerSummaryData{
				Contact: Contact{
					PrimaryContactId:    currentCustomer.Id,
					Emails:              []string{*currentCustomer.Email},
					PhoneNumbers:        []string{*currentCustomer.PhoneNumber},
					SecondaryContactIds: []int{},
				},
			}

			utils.SendJSONMessageResponse(w, customerSummary, http.StatusOK)
			return
		}
	} else if len(allMatchingCustomers) == 0 {
		////////
		fmt.Println("bruh2")

	} else if len(allMatchingCustomers) == 2 {
		////
		fmt.Println("bruh 2 customers")

		// case 1: one is primary customer & the other is secondary
		// case 2: both are primary ==> special case

		// find number of matching customers
		primaryCustomerCount := 0
		for i := 0; i < 2; i++ {
			if allMatchingCustomers[i].LinkPrecedence == Primary {
				primaryCustomerCount++
			}
		}

		if primaryCustomerCount == 1 {
			customerSummary := CustomerSummaryData{}
			// add details of primary contact first. still would be O(n) time complexity
			for i := 0; i < len(allMatchingCustomers); i++ {
				c := allMatchingCustomers[i]
				if c.LinkPrecedence == Primary {
					customerSummary.Contact.PrimaryContactId = c.Id
					customerSummary.Contact.Emails = append(
						customerSummary.Contact.Emails,
						*c.Email,
					)
					customerSummary.Contact.PhoneNumbers = append(
						customerSummary.Contact.PhoneNumbers,
						*c.PhoneNumber,
					)
				}
			}

			// now add secondary stuff
			for i := 0; i < len(allMatchingCustomers); i++ {
				c := allMatchingCustomers[i]
				if c.LinkPrecedence == Secondary {
					customerSummary.Contact.Emails = append(
						customerSummary.Contact.Emails,
						*c.Email,
					)
					customerSummary.Contact.PhoneNumbers = append(
						customerSummary.Contact.PhoneNumbers,
						*c.PhoneNumber,
					)
					customerSummary.Contact.SecondaryContactIds = append(
						customerSummary.Contact.SecondaryContactIds,
						c.Id,
					)
				}
			}

			// return the customer summary data

			utils.SendJSONMessageResponse(w, customerSummary, http.StatusOK)
			return

		} else if primaryCustomerCount == 2 {
			fmt.Println("bruh functionality not implemented")
		} else {
			// panic
			fmt.Println("bruh Unexpected program state")
		}

	} else if len(allMatchingCustomers) > 2 {
		fmt.Println("bruh matching more than 2: ", len(allMatchingCustomers))
		customerSummary := CustomerSummaryData{}
		// add details of primary contact first. still would be O(n) time complexity
		for i := 0; i < len(allMatchingCustomers); i++ {
			c := allMatchingCustomers[i]
			if c.LinkPrecedence == Primary {
				customerSummary.Contact.PrimaryContactId = c.Id
				customerSummary.Contact.Emails = append(
					customerSummary.Contact.Emails,
					*c.Email,
				)
				customerSummary.Contact.PhoneNumbers = append(
					customerSummary.Contact.PhoneNumbers,
					*c.PhoneNumber,
				)
			}
		}

		// now add secondary stuff
		for i := 0; i < len(allMatchingCustomers); i++ {
			c := allMatchingCustomers[i]
			if c.LinkPrecedence == Secondary {
				customerSummary.Contact.Emails = append(
					customerSummary.Contact.Emails,
					*c.Email,
				)
				customerSummary.Contact.PhoneNumbers = append(
					customerSummary.Contact.PhoneNumbers,
					*c.PhoneNumber,
				)
				customerSummary.Contact.SecondaryContactIds = append(
					customerSummary.Contact.SecondaryContactIds,
					c.Id,
				)
			}
		}

		// return the customer summary data

		utils.SendJSONMessageResponse(w, customerSummary, http.StatusOK)
		return

	}

	// primaryCustomerCount, err := getPrimaryCustomerCount(db, *phoneNumber, *email)
	// if primaryCustomerCount < 1 {
	// 	// primary customer doesnt exist
	// 	// add a new entry to the database
	// 	err := insertNewCustomer(db, *phoneNumber, *email)
	// 	if err != nil {
	// 		message := fmt.Sprintf("Failed to add new customer into database: %v", err)
	// 		statusCode := http.StatusInternalServerError
	// 		utils.SendErrorResponse(w, message, statusCode)
	// 		return
	// 	}
	// } else {
	// 	// in this block we are sure that a primary user exists

	// 	if primaryCustomerCount == 1 {
	// 		customerIsPrimary, err := isPrimaryCustomer(db, *phoneNumber, *email)
	// 		if err != nil {
	// 			message := fmt.Sprintf("SQL query failed: %v", err)
	// 			statusCode := http.StatusInternalServerError
	// 			utils.SendErrorResponse(w, message, statusCode)
	// 			return
	// 		}
	// 		if customerIsPrimary {
	// 			// single primary customer exists
	// 			// input data belongs to the primary customer
	// 			// thus we just need to return the primary customer itself

	// 			// need to fetch user details?
	// 			{
	// 			contact:
	// 				{

	// 				}
	// 			}

	// 		} else {
	// 			// single primary customer exists
	// 			// and the input data does not belong to the primary customer
	// 			// thus we need to add a secondary user
	// 			// return both primary and newly added secondary users
	// 		}

	// 	} else {
	// 		// multiple primary customers exist
	// 		// thus we need to link both the primary customers
	// 		// the oldest record will be the primary
	// 		// new record will be the secondary
	// 		// but they will be linked

	// 	}

	// }
}
