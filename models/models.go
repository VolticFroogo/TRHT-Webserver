package models

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Token lifetimes
const (
	// AuthTokenValidTime is the lifetime of an auth token.
	AuthTokenValidTime = time.Minute * 15
	// RefreshTokenValidTime is the lifetime of a refresh token.
	RefreshTokenValidTime = time.Hour * 72
)

// Privileges
const (
	PrivNone = iota
	PrivAdmin
	PrivSuperAdmin
)

// Slide is a slide for the slides on the index page.
type Slide struct {
	ID                        int
	Image, Title, Description string
}

// Slides is an array of slides for the slides on the index page.
type Slides []Slide

// MenuItem is an item for the menu on the index page.
type MenuItem struct {
	ID                       int
	Name, Description, Price string
}

// Menu is an array of MenuItems for the menu on the index page.
type Menu []MenuItem

// MenuItemEdit is the struct recieved by an admin when they change a menu item.
type MenuItemEdit struct {
	ID                                   int
	CsrfSecret, Name, Description, Price string
}

// UserEdit is the struct recieved by an admin when they update a user.
type UserEdit struct {
	ID, Privileges                            int
	CsrfSecret, Email, Password, Fname, Lname string
}

// ContactMessage is the struct for a message on the admin page.
type ContactMessage struct {
	ID                   int
	Name, Email, Message string
}

// ContactMessages is an array of ContactMessage for the admin page.
type ContactMessages []ContactMessage

// ContactMessageEdit is the struct recieved by an admin when they delete a contact message.
type ContactMessageEdit struct {
	ID         int
	CsrfSecret string
}

// SlideEdit is the struct recieved by an admin when they delete a slide.
type SlideEdit struct {
	ID         int
	CsrfSecret string
}

// User is a user retrieved from a Database.
type User struct {
	UUID, Priv                                int
	Email, Password, Fname, Lname, CreateTime string
}

// Users is an array of User for the admin page.
type Users []User

// TokenClaims are the claims in a token.
type TokenClaims struct {
	jwt.StandardClaims
	CSRF string `json:"csrf"`
}

// TemplateVariables is the struct used when executing a template.
type TemplateVariables struct {
	CsrfSecret      string
	User            User
	Slides          Slides
	Menu            Menu
	ContactMessages ContactMessages
	Users           Users
	SuperAdmin      bool
}

// AJAXData is the struct used with the AJAX middleware.
type AJAXData struct {
	CsrfSecret string
}

// JTI is the struct used for JTIs in the DB.
type JTI struct {
	ID     int
	Expiry int64
	JTI    string
}
