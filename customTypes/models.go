package customtypes

import "database/sql"

// Custom data types

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
