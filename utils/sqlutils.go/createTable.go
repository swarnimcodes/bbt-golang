package sqlutils

import "database/sql"

func CreateCustomerTable(db *sql.DB) error {
	tableCreationQuery := `
		CREATE TABLE IF NOT EXISTS Contact (
			id 				INTEGER PRIMARY KEY AUTOINCREMENT,
			PhoneNumber 	TEXT NOT NULL,
			email 			TEXT NOT NULL,
			linkedId 		INTEGER,
			linkPrecedence 	TEXT NOT NULL,
			createdAt 		TEXT NOT NULL,  -- ISO 8601
			updatedAt 		TEXT NOT NULL,
			deletedAt 		TEXT 
		)
	`
	_, err := db.Exec(tableCreationQuery)
	return err
}
