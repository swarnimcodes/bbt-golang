package customer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	customtypes "github.com/swarnimcodes/bitespeed-backend-task/customTypes"
	"github.com/swarnimcodes/bitespeed-backend-task/utils"
	"github.com/swarnimcodes/bitespeed-backend-task/utils/sqlutils.go"
)

func prependToSlice[T any](slice []T, element T) []T {
	return append([]T{element}, slice...)
}

func isElementDuplicate[T comparable](slice []T, element T) bool {
	if len(slice) == 0 {
		return false
	}
	for i := 0; i < len(slice); i++ {
		if slice[i] == element {
			return true
		}
	}
	return false
}

// whenever this function is called make sure that only one primary account exists
func sendContactSummary(w http.ResponseWriter, familyMembers []customtypes.Customer) error {
	var primaryContactId int
	var emails []string
	var secondaryContactIds []int
	var phoneNumbers []string
	for i := 0; i < len(familyMembers); i++ {
		if familyMembers[i].LinkPrecedence == customtypes.Primary {
			primaryContactId = familyMembers[i].Id
			if !isElementDuplicate(emails, *familyMembers[i].Email) {
				emails = prependToSlice(emails, *familyMembers[i].Email)
			}
			if !isElementDuplicate(phoneNumbers, *familyMembers[i].PhoneNumber) {
				phoneNumbers = prependToSlice(phoneNumbers, *familyMembers[i].PhoneNumber)
			}

		} else if familyMembers[i].LinkPrecedence == customtypes.Secondary {
			primaryContactId = int(familyMembers[i].LinkedId.Int32)

			if !isElementDuplicate(emails, *familyMembers[i].Email) {
				emails = append(emails, *familyMembers[i].Email)
			}
			if !isElementDuplicate(phoneNumbers, *familyMembers[i].PhoneNumber) {
				phoneNumbers = append(phoneNumbers, *familyMembers[i].PhoneNumber)
			}
			secondaryContactIds = append(secondaryContactIds, familyMembers[i].Id)

		} else {
			fmt.Println("we're fucked")
			err := errors.New("could not fetch primary contact id")
			return err
		}
	}

	customerSummar := customtypes.CustomerSummaryData{
		Contact: customtypes.Contact{
			PrimaryContactId:    primaryContactId,
			Emails:              emails,
			PhoneNumbers:        phoneNumbers,
			SecondaryContactIds: secondaryContactIds,
		},
	}

	utils.SendJSONMessageResponse(w, customerSummar, http.StatusOK)
	return nil
}

func printCustomers(customers []customtypes.Customer) {
	for _, member := range customers {
		fmt.Printf("ID: %d\n", member.Id)
		fmt.Printf("Email: %s\n", *member.Email)
		fmt.Printf("Phone Number: %s\n", *member.PhoneNumber)
		fmt.Printf("Linked ID: %v\n", member.LinkedId.Int32)
		fmt.Printf("Link Precedence: %v\n", member.LinkPrecedence)
		fmt.Printf("Created At: %s\n", member.CreatedAt)
		fmt.Printf("Updated At: %s\n", member.UpdatedAt)
		fmt.Printf("Deleted At: %s\n", member.DeletedAt.String)
		fmt.Println("-------------------------")
	}
}

func IdentifyCustomer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hello?")
	db, err := utils.GetDatabaseConnection()
	if err != nil {
		errMsg := fmt.Sprintf("could not connect to the database: %v", err)
		utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Create user table if it doesn't exist
	if err := sqlutils.CreateCustomerTable(db); err != nil {
		message := fmt.Sprintf("Failed to create user table in the database: %v", err)
		statusCode := http.StatusInternalServerError
		utils.SendErrorResponse(w, message, statusCode)
		return
	}

	// parse request body into User struct
	// var user User
	var customer customtypes.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		message := fmt.Sprintf("Invalid JSON request input payload: %v", err)
		statusCode := http.StatusBadRequest
		utils.SendErrorResponse(w, message, statusCode)
		return
	}

	// TODO: check if phone number or email exists in database.
	var phoneNumber, email string
	if customer.PhoneNumber != nil {
		phoneNumber = *customer.PhoneNumber
	}
	if customer.Email != nil {
		email = *customer.Email
	}

	familyMembers, err := sqlutils.GetRelatedMatches(db, phoneNumber, email)
	if err != nil {
		errMsg := fmt.Sprintf("could not get related matches from the database: %v", err)
		utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}
	fmt.Println("First Pass: ")
	printCustomers(familyMembers)

	// if length of familyMembers == 0 add a new primary account
	if len(familyMembers) == 0 {
		isInformationNew, err := sqlutils.NewInformationReceived(db, phoneNumber, email)
		if err != nil {
			errMsg := fmt.Sprintf("error in determining if information received is new: %v", err)
			utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}

		if isInformationNew {
			err := sqlutils.InsertNewPrimaryCustomer(db, phoneNumber, email)
			if err != nil {
				errMsg := fmt.Sprintf("could not insert new contact into the database: %v", err)
				utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
				return
			}
		}

		// now ideally the `GetRelatedMatches` function should return the newly
		// inserted entry itself
		newFamilyMembers, err := sqlutils.GetRelatedMatches(db, phoneNumber, email)
		if err != nil {
			errMsg := fmt.Sprintf("could not get related matches from the database: %v", err)
			utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}

		// send properly formatted response back to the api caller
		if err := sendContactSummary(w, newFamilyMembers); err != nil {
			errMsg := fmt.Sprintf("could not send proper response back: %v", err)
			utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}

	}

	// here len of familyMembers != 0 and we have all related familyMembers
	// find number of primary accounts
	var primaryAccountIds []int
	var secondaryAccountIds []int
	for i := 0; i < len(familyMembers); i++ {
		if familyMembers[i].LinkPrecedence == customtypes.Primary {
			primaryAccountIds = append(primaryAccountIds, familyMembers[i].Id)
		} else if familyMembers[i].LinkPrecedence == customtypes.Secondary {
			secondaryAccountIds = append(secondaryAccountIds, familyMembers[i].Id)
		}
	}

	primaryAccountCount := len(primaryAccountIds)

	if primaryAccountCount == 1 {
		// insert a secondary account
		isInformationNew, err := sqlutils.NewInformationReceived(db, phoneNumber, email)
		if err != nil {
			errMsg := fmt.Sprintf("error in determining if information received is new: %v", err)
			utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}
		if isInformationNew {
			if err := sqlutils.InsertNewSecondaryCustomer(db, phoneNumber, email, primaryAccountIds[0]); err != nil {
				errMsg := fmt.Sprintf("could not insert secondary account: %v", err)
				utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
				return
			}
		}

		// TODO: unnecessary DB call if information received is not new?
		newFamilyMembers, err := sqlutils.GetRelatedMatches(db, phoneNumber, email)
		if err != nil {
			errMsg := fmt.Sprintf("could not get related matches from the database: %v", err)
			utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}

		if err := sendContactSummary(w, newFamilyMembers); err != nil {
			errMsg := fmt.Sprintf("could not send proper response back: %v", err)
			utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}

	} else if primaryAccountCount == 2 {
		// 2 primary accounts
		// we need to find the oldest primary account and keep that primary
		// the second primary account needs to be linked to the new primary account
		// all the children of the second primary account also need to be linked with the 1st primary account

		var primaryId int
		var accountIdToChange int

		if primaryAccountIds[0] < primaryAccountIds[1] {
			primaryId = primaryAccountIds[0]
			accountIdToChange = primaryAccountIds[1]
		} else {
			primaryId = primaryAccountIds[1]
			accountIdToChange = primaryAccountIds[0]
		}

		// update the newer primary account to be secondary
		if err := sqlutils.UpdateMembersLinkedId(db, primaryId, []int{accountIdToChange}); err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// update the children of the newer primary account
		// make the linkedId = primaryId
		accountIdsToChange := append(secondaryAccountIds, accountIdToChange)
		if err := sqlutils.UpdateMembersLinkedId(db, primaryId, accountIdsToChange); err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// get family again?
		newFamilyMembers, err := sqlutils.GetRelatedMatches(db, phoneNumber, email)
		if err != nil {
			errMsg := fmt.Sprintf("could not get related matches from the database: %v", err)
			utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}

		if err := sendContactSummary(w, newFamilyMembers); err != nil {
			errMsg := fmt.Sprintf("could not send proper response back: %v", err)
			utils.SendErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}
		return

	} else {
		fmt.Println("we're fucked")
	}

	// =========================> <==================================

}
