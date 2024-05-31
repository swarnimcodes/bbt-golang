package customer

import (
	"fmt"
	"net/http"

	customtypes "github.com/swarnimcodes/bitespeed-backend-task/customTypes"
	"github.com/swarnimcodes/bitespeed-backend-task/utils"
	"github.com/swarnimcodes/bitespeed-backend-task/utils/sqlutils.go"
)

func GetAllCustomers(w http.ResponseWriter, r *http.Request) {
	// get shared database connection
	db, err := utils.GetDatabaseConnection()
	if err != nil {
		errMsg := fmt.Sprintf("could not connect to the database: %v", err)
		utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	err = sqlutils.CreateCustomerTable(db)
	if err != nil {
		errMsg := fmt.Sprintf("could not create table: %v", err)
		utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT * FROM Contact")
	if err != nil {
		errMsg := fmt.Sprintf("could not fetch data from database: %v", err)
		utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}
	var customers []customtypes.Customer

	for rows.Next() {
		var customer customtypes.Customer
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
			errMsg := fmt.Sprintf("could not fetch row data: %v", err)
			utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}
		customers = append(customers, customer)
	}

	utils.SendJSONMessageResponse(w, customers, http.StatusOK)
}
