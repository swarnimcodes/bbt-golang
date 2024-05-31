package sqlutils

import (
	"database/sql"
	"fmt"
	"strings"

	customtypes "github.com/swarnimcodes/bitespeed-backend-task/customTypes"
)

func findFamily(db *sql.DB, ids []int) ([]customtypes.Customer, error) {
	placeholders := make([]string, len(ids))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(`
	SELECT id, phoneNumber, email, linkedId, linkPrecedence, createdAt, updatedAt, deletedAt
	FROM Contact
	WHERE id IN (%s)
	OR linkedId IN (%s)
	`, strings.Join(placeholders, ", "), strings.Join(placeholders, ", "))

	fmt.Printf("Query to find family members: %s", query)

	args := make([]interface{}, len(ids)*2)
	for i, id := range ids {
		args[i] = id
		args[i+len(ids)] = id
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var familyMembers []customtypes.Customer

	for rows.Next() {
		var c customtypes.Customer
		err := rows.Scan(
			&c.Id,
			&c.PhoneNumber,
			&c.Email,
			&c.LinkedId,
			&c.LinkPrecedence,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		familyMembers = append(familyMembers, c)
	}

	return familyMembers, nil
}

func GetRelatedMatches(db *sql.DB, phoneNumber, email string) ([]customtypes.Customer, error) {
	// step 1: find all the accounts where either the email or the phone number match
	var customers []customtypes.Customer
	query := `
		SELECT *
		FROM Contact
		WHERE phoneNumber = ? OR email = ?
	`
	rows, err := db.Query(query, phoneNumber, email)
	if err != nil {
		return []customtypes.Customer{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var c customtypes.Customer
		if err := rows.Scan(
			&c.Id,
			&c.PhoneNumber,
			&c.Email,
			&c.LinkedId,
			&c.LinkPrecedence,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		); err != nil {
			return []customtypes.Customer{}, err
		}
		customers = append(customers, c)
	}

	if len(customers) == 0 {
		return []customtypes.Customer{}, nil // ==> the caller should insert new user in this case
	}

	// step 2: find primary account ids that are ancestors
	// here we have all the accounts where either field match
	// loop over them and find all the linkedIds and primary accountIds
	var primaryAccountIds []int
	for i := 0; i < len(customers); i++ {
		accountType := customers[i].LinkPrecedence
		if accountType == customtypes.Primary {
			// add its id to the slice
			primaryAccountIds = append(primaryAccountIds, customers[i].Id)
		} else if accountType == customtypes.Secondary {
			// add its linked id to the slice
			if customers[i].LinkedId.Valid {
				primaryAccountIds = append(primaryAccountIds, int(customers[i].LinkedId.Int32))
			} else {
				fmt.Println("we're fucked")
			}

		} else {
			fmt.Println("we're fucked")
		}
	}

	// now we have all the ancestors
	// step 3: find all children including ancestors
	familyMembers, err := findFamily(db, primaryAccountIds)
	if err != nil {
		return nil, err
	}
	// now we have all the related accounts!
	// job of this function is done
	return familyMembers, nil
}
