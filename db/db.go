package db

import (
	"database/sql"
	"time"

	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/db/dbCredentials"
	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/helpers"
	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/models"
	_ "github.com/go-sql-driver/mysql" // Necessary for connecting to MySQL.
)

/*
	Structs and variables
*/

var (
	refreshTokens map[string]int64
	db            *sql.DB
	// Slides is a struct for the slides.
	Slides models.Slides
	// Menu is a struct for the slides.
	Menu models.Menu
	// ContactMessages is a struct for the admin contact messages.
	ContactMessages models.ContactMessages
)

// InitDB initializes the Database.
func InitDB() (err error) {
	refreshTokens = make(map[string]int64)
	db, err = sql.Open(dbCredentials.Type, dbCredentials.ConnString)
	UpdateSlides()
	UpdateMenu()
	UpdateContactMessages()
	return
}

/*
	Non-MySQL DataBase related functions
*/

// StoreRefreshToken generates, stores and then returns a jti.
func StoreRefreshToken() (jti string, err error) {
	jti, err = helpers.GenerateRandomString(32)
	if err != nil {
		return jti, err
	}

	// Check to make sure our jti is unique.
	for refreshTokens[jti] != 0 {
		jti, err = helpers.GenerateRandomString(32)
		if err != nil {
			return jti, err
		}
	}

	refreshTokens[jti] = time.Now().Add(models.RefreshTokenValidTime).Unix()

	return jti, err
}

// CheckJti returns the validity of a jti.
func CheckJti(jti string) (valid bool) {
	if refreshTokens[jti] > time.Now().Unix() { // Check if token has expired.
		return true // Token is valid.
	}

	delete(refreshTokens, jti)
	return false // Token is invalid.
}

func jtiGarbageCollector(quit chan bool) {
	ticker := time.NewTicker(5 * time.Minute) // Tick every five minutes.
	for {
		<-ticker.C                                     // Tick: run garbage collector.
		var jti string                                 // Make a string to store a tokenID.
		for refreshTokenRange := range refreshTokens { // Make a range of all tokens.
			jti = refreshTokenRange // Set a tokenID from range.
			CheckJti(jti)           // Check if token is valid if not it's deleted.
		}
	}
}

/*
	MySQL DataBase related functions
*/

// GetUserFromID retrieves a user from the MySQL database.
func GetUserFromID(uuid int) (user models.User, err error) {
	rows, err := db.Query("SELECT email, password, fname, lname, create_time FROM users WHERE uuid = ?", uuid) // Query DB for id and password from username.
	if err != nil {
		return
	}

	defer rows.Close()

	user.UUID = uuid
	for rows.Next() {
		err = rows.Scan(&user.Email, &user.Password, &user.Fname, &user.Lname, &user.CreateTime) // Scan data from query.
		if err != nil {
			return
		}
	}

	return
}

// GetUserFromEmail retrieves a user's ID from the MySQL database.
func GetUserFromEmail(email string) (user models.User, err error) {
	rows, err := db.Query("SELECT uuid, password, fname, lname, create_time FROM users WHERE email = ?", email) // Query DB for id and password from username.
	if err != nil {
		return
	}

	defer rows.Close()

	user.Email = email
	for rows.Next() {
		err = rows.Scan(&user.UUID, &user.Password, &user.Fname, &user.Lname, &user.CreateTime) // Scan data from query.
		if err != nil {
			return
		}
	}

	return
}

// UpdateSlides updates the slides by querying the MySQL DataBase.
func UpdateSlides() (err error) {
	rows, err := db.Query("SELECT id, image, title, description FROM slides") // Query DB for id and password from username.
	if err != nil {
		return
	}

	defer rows.Close()

	slides := models.Slides{} // Create struct to store slides in.
	slide := models.Slide{}   // Create struct to store a slide in.
	for rows.Next() {
		err = rows.Scan(&slide.ID, &slide.Image, &slide.Title, &slide.Description) // Scan data from query.
		if err != nil {
			return
		}

		slides = append(slides, slide) // Append just read slide into the slides.
	}

	Slides = slides // Replace the old slides with the newly read struct.
	return
}

// UpdateMenu updates the menu by querying the MySQL DataBase.
func UpdateMenu() (err error) {
	rows, err := db.Query("SELECT id, name, description, price FROM menu") // Query DB for id and password from username
	if err != nil {
		return
	}

	defer rows.Close()

	menu := models.Menu{}         // Create struct to store slides in.
	menuItem := models.MenuItem{} // Create struct to store a slide in.
	for rows.Next() {
		err = rows.Scan(&menuItem.ID, &menuItem.Name, &menuItem.Description, &menuItem.Price) // Scan data from query.
		if err != nil {
			return
		}

		menu = append(menu, menuItem) // Append just read slide into the slides.
	}

	Menu = menu // Replace the old menu with the newly read struct.
	return
}

// UpdateContactMessages updates the messages by querying the MySQL DataBase.
func UpdateContactMessages() (err error) {
	rows, err := db.Query("SELECT id, name, email, message FROM contact") // Query DB for id and password from username
	if err != nil {
		return
	}

	defer rows.Close()

	contactMessages := models.ContactMessages{} // Create struct to store slides in.
	contactMessage := models.ContactMessage{}   // Create struct to store a slide in.
	for rows.Next() {
		err = rows.Scan(&contactMessage.ID, &contactMessage.Name, &contactMessage.Email, &contactMessage.Message) // Scan data from query.
		if err != nil {
			return
		}

		contactMessages = append(contactMessages, contactMessage) // Append just read slide into the slides.
	}

	ContactMessages = contactMessages // Replace the old menu with the newly read struct.
	return
}

// EditMenuItem update's a menu item.
func EditMenuItem(ID int, Name, Description, Price string) (err error) {
	_, err = db.Exec("UPDATE menu SET name=?, description=?, price=? WHERE id=?", Name, Description, Price, ID)
	if err != nil {
		return
	}

	err = UpdateMenu()
	return
}

// NewMenuItem creates a new menu item.
func NewMenuItem(Name, Description, Price string) (id int, err error) {
	_, err = db.Exec("INSERT INTO menu (name, description, price) VALUES (?, ?, ?)", Name, Description, Price)
	if err != nil {
		return
	}

	rows, err := db.Query("SELECT id FROM menu WHERE name = ? AND description = ? AND price = ?", Name, Description, Price) // Query DB for id and password from username
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	err = rows.Scan(&id)
	if err != nil {
		return
	}

	err = UpdateMenu()
	return
}

// DeleteMenuItem deletes a menu item.
func DeleteMenuItem(ID int) (err error) {
	_, err = db.Exec("DELETE FROM menu WHERE id=?", ID)
	if err != nil {
		return
	}

	err = UpdateMenu()
	return
}

// NewContactMessage adds a new contact message.
func NewContactMessage(Name, Email, Message string) (err error) {
	_, err = db.Exec("INSERT INTO contact (name, email, message) VALUES (?, ?, ?)", Name, Email, Message)
	if err != nil {
		return
	}

	err = UpdateContactMessages()
	return
}

// DeleteContactMessage deletes a menu item.
func DeleteContactMessage(ID int) (err error) {
	_, err = db.Exec("DELETE FROM contact WHERE id=?", ID)
	if err != nil {
		return
	}

	err = UpdateContactMessages()
	return
}