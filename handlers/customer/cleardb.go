package customer

import (
	"fmt"
	"net/http"

	"github.com/swarnimcodes/bitespeed-backend-task/utils"
)

func ClearDb(w http.ResponseWriter, r *http.Request) {
	db, err := utils.GetDatabaseConnection()
	if err != nil {
		errMsg := fmt.Sprintf("could not connect to the database: %v", err)
		utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	query := `DROP TABLE Contact`

	_, err = db.Exec(query)
	if err != nil {
		errMsg := fmt.Sprintf("error while dropping table: %v", err)
		utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	message := "table dropped successfully"
	utils.SendMessageResponse(w, message, http.StatusOK)
	return
}
