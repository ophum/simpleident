package models

import "github.com/google/uuid"

type Oauth2Client struct {
	Model
	Name        string
	Description string
	CallbackURL string
	// has many
	ClientSecrets []*Oauth2ClientSecret
}

type Oauth2ClientSecret struct {
	Model
	// belongs to
	Oauth2ClientID uuid.UUID
	Oauth2Client   *Oauth2Client

	Secret string
}

type Oauth2Code struct {
	Model
	Oauth2ClientID uuid.UUID
	Code           string
	AccountID      uuid.UUID
}

type Oauth2Token struct {
	Model
	Oauth2ClientID uuid.UUID
	Token          string
	AccountID      uuid.UUID
	Account        *Account
}
