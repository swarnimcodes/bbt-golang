package sqlutils

import "database/sql"

func emailExists(db *sql.DB, email string) (bool, error) {
	if email == "" {
		return true, nil
	}
	query := `
		SELECT EXISTS(
			SELECT 1 FROM Contact
			WHERE email = ?
			AND deletedAt IS NULL
		)
	`
	var exists bool
	err := db.QueryRow(query, email).Scan(&exists)
	return exists, err
}

func phoneNumberExists(db *sql.DB, phoneNumber string) (bool, error) {
	if phoneNumber == "" {
		return true, nil
	}
	query := `
		SELECT EXISTS(
			SELECT 1 FROM Contact
			WHERE phoneNumber = ?
			AND deletedAt IS NULL
		)
	`
	var exists bool
	err := db.QueryRow(query, phoneNumber).Scan(&exists)
	return exists, err
}

func NewInformationReceived(db *sql.DB, phoneNumber, email string) (bool, error) {
	pE, err := phoneNumberExists(db, phoneNumber)
	if err != nil {
		return false, err
	}

	eE, err := emailExists(db, email)
	if err != nil {
		return false, err
	}

	if pE && eE {
		return false, nil
	} else {
		return true, nil
	}
}
