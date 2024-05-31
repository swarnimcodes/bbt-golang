package sqlutils

import (
	"database/sql"
	"fmt"
	"strings"

	customtypes "github.com/swarnimcodes/bitespeed-backend-task/customTypes"
)

func UpdateMembersLinkedId(db *sql.DB, primaryId int, accountIds []int) error {
	// make the slice of integers to string
	ids := make([]string, len(accountIds))
	for i, id := range accountIds {
		ids[i] = fmt.Sprintf("%d", id)
	}
	idList := strings.Join(ids, ", ")

	query := fmt.Sprintf(`
		UPDATE Contact
		SET linkedId = ?, linkPrecedence = ?, updatedAt = datetime('now')
		WHERE id IN (%s)
	`, idList)

	_, err := db.Exec(query, primaryId, customtypes.Secondary)
	if err != nil {
		return fmt.Errorf("failed to update member(s) with id(s) %s: %v", idList, err)
	}
	return nil
}
