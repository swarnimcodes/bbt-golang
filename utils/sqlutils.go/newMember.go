package sqlutils

import "database/sql"

func InsertNewPrimaryCustomer(db *sql.DB, phoneNumber, email string) error {
	query := `
		INSERT INTO Contact (phoneNumber, email, linkedId, linkPrecedence, createdAt, updatedAt, deletedAt)
		VALUES (?, ?, NULL, 'primary', datetime('now'), datetime('now'), NULL)
	`
	_, err := db.Exec(query, phoneNumber, email)
	return err
}

func InsertNewSecondaryCustomer(db *sql.DB, phoneNumber string, email string, relatedPrimaryId int) error {
	query := `
		INSERT INTO Contact (phoneNumber, email, linkedId, linkPrecedence, createdAt, updatedAt, deletedAt)
		VALUES (?, ?, ?, 'secondary',  datetime('now'), datetime('now'), NULL)
	`
	_, err := db.Exec(query, phoneNumber, email, relatedPrimaryId)
	return err
}
