package user

import "database/sql"

type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
	Location  string `json:"location"`
}

// Customer type to parse json payload
// email and phone number can be null
type LinkPrecedence string

const (
	Primary   LinkPrecedence = "primary"
	Secondary LinkPrecedence = "secondary"
)

type Customer struct {
	Email          *string        `json:"email"`
	PhoneNumber    *string        `json:"phoneNumber"`
	Id             int            `json:"id"`
	LinkedId       sql.NullInt32  `json:"linkedId"`
	LinkPrecedence LinkPrecedence `json:"linkPrecedence"`
	CreatedAt      string         `json:"createdAt"`
	UpdatedAt      string         `json:"updatedAt"`
	DeletedAt      sql.NullString `json:"deletedAt"`
}

// Response type
type Contact struct {
	PrimaryContactId    int      `json:"primaryContactId"`
	Emails              []string `json:"emails"`
	PhoneNumbers        []string `json:"phoneNumbers"`
	SecondaryContactIds []int    `json:"secondaryContactIds"`
}
type CustomerSummaryData struct {
	Contact Contact `json:"contact"`
}
