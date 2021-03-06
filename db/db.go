package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/VolticFroogo/TRHT-Webserver/db/dbCredentials"
	"github.com/VolticFroogo/TRHT-Webserver/helpers"
	"github.com/VolticFroogo/TRHT-Webserver/models"
	_ "github.com/go-sql-driver/mysql" // Necessary for connecting to MySQL.
)

/*
	Structs and variables
*/

var (
	db *sql.DB
	// Slides is a struct for the slides.
	Slides models.Slides
	// Menu is a struct for the slides.
	Menu models.Menu
	// ContactMessages is a struct for the admin contact messages.
	ContactMessages models.ContactMessages
	// Users is a struct for the admin Users.
	Users models.Users
)

// InitDB initializes the Database.
func InitDB() (err error) {
	db, err = sql.Open(dbCredentials.Type, dbCredentials.ConnString)
	UpdateSlides()
	UpdateMenu()
	UpdateContactMessages()
	UpdateUsers()
	go jtiGarbageCollector()
	go autoUpdate()
	return
}

/*
	For some reason (probably garbage collection) this program
	loses all of it's data displayed everywhere so this compensates
	for this by every so often updating itself automatically.
*/
func autoUpdate() {
	ticker := time.NewTicker(time.Hour) // Tick every hour.
	for {
		<-ticker.C

		UpdateSlides()
		UpdateMenu()
		UpdateContactMessages()
		UpdateUsers()
	}
}

/*
	MySQL DataBase related functions
*/

// StoreRefreshToken generates, stores and then returns a JTI.
func StoreRefreshToken() (jti models.JTI, err error) {
	// No need to duplication check as the JTI takes input from time and are unique.
	jti.JTI, err = helpers.GenerateRandomString(32)
	if err != nil {
		return
	}

	jti.Expiry = time.Now().Add(models.RefreshTokenValidTime).Unix()

	_, err = db.Exec("INSERT INTO jti (jti, expiry) VALUES (?, ?)", jti.JTI, jti.Expiry)
	if err != nil {
		return
	}

	rows, err := db.Query("SELECT id FROM jti WHERE jti=? AND expiry=?", jti.JTI, jti.Expiry)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	err = rows.Scan(&jti.ID) // Scan data from query.
	return
}

// GetJTI takes a JTI string and returns the JTI struct.
func GetJTI(jti string) (jtiStruct models.JTI, err error) {
	rows, err := db.Query("SELECT id, expiry FROM jti WHERE jti=?", jti)
	if err != nil {
		return
	}

	defer rows.Close()

	jtiStruct.JTI = jti
	rows.Next()
	err = rows.Scan(&jtiStruct.ID, &jtiStruct.Expiry) // Scan data from query.
	return
}

// CheckJTI returns the validity of a JTI.
func CheckJTI(jti models.JTI) (valid bool, err error) {
	if jti.Expiry > time.Now().Unix() { // Check if token has expired.
		return true, nil // Token is valid.
	}

	_, err = db.Exec("DELETE FROM jti WHERE id=?", jti.ID)
	if err != nil {
		return false, err
	}

	return false, nil // Token is invalid.
}

// DeleteJTI deletes a JTI based on a jti key.
func DeleteJTI(jti string) (err error) {
	_, err = db.Exec("DELETE FROM jti WHERE jti=?", jti)
	return
}

func jtiGarbageCollector() {
	ticker := time.NewTicker(5 * time.Minute) // Tick every five minutes.
	for {
		<-ticker.C
		rows, err := db.Query("SELECT id, jti, expiry FROM jti")
		if err != nil {
			log.Printf("Error querying JTI DB in JTI garbage collector: %v", err)
			return
		}

		defer rows.Close()

		jti := models.JTI{} // Create struct to store a JTI in.
		for rows.Next() {
			err = rows.Scan(&jti.ID, &jti.JTI, &jti.Expiry) // Scan data from query.
			if err != nil {
				log.Printf("Error scanning rows in JTI garbage collector: %v", err)
				return
			}

			_, err := CheckJTI(jti)
			if err != nil {
				log.Printf("Error checking in JTI garbage collector: %v", err)
				return
			}
		}
	}
}

// GetUserFromID retrieves a user from the MySQL database.
func GetUserFromID(uuid int) (user models.User, err error) {
	rows, err := db.Query("SELECT email, password, fname, lname, priv, create_time FROM users WHERE uuid=?", uuid)
	if err != nil {
		return
	}

	defer rows.Close()

	user.UUID = uuid
	for rows.Next() {
		err = rows.Scan(&user.Email, &user.Password, &user.Fname, &user.Lname, &user.Priv, &user.CreateTime) // Scan data from query.
		if err != nil {
			return
		}
	}

	return
}

// GetUserFromEmail retrieves a user's ID from the MySQL database.
func GetUserFromEmail(email string) (user models.User, err error) {
	rows, err := db.Query("SELECT uuid, password, fname, lname, priv, create_time FROM users WHERE email=?", email)
	if err != nil {
		return
	}

	defer rows.Close()

	user.Email = email
	for rows.Next() {
		err = rows.Scan(&user.UUID, &user.Password, &user.Fname, &user.Lname, &user.Priv, &user.CreateTime) // Scan data from query.
		if err != nil {
			return
		}
	}

	return
}

// UpdateSlides updates the slides by querying the MySQL DataBase.
func UpdateSlides() (err error) {
	rows, err := db.Query("SELECT id, image, title, description FROM slides")
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
	rows, err := db.Query("SELECT id, name, description, price FROM menu")
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
	rows, err := db.Query("SELECT id, name, email, message FROM contact")
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

// UpdateUsers updates the users by querying the MySQL DataBase.
func UpdateUsers() (err error) {
	rows, err := db.Query("SELECT uuid, email, fname, lname, priv FROM users")
	if err != nil {
		return
	}

	defer rows.Close()

	users := models.Users{} // Create struct to store slides in.
	user := models.User{}   // Create struct to store a slide in.
	for rows.Next() {
		err = rows.Scan(&user.UUID, &user.Email, &user.Fname, &user.Lname, &user.Priv) // Scan data from query.
		if err != nil {
			return
		}

		users = append(users, user) // Append just read slide into the slides.
	}

	Users = users // Replace the old menu with the newly read struct.
	return
}

// NewSlide creates a new slide.
func NewSlide(Title, Description, Image string) (id int, err error) {
	_, err = db.Exec("INSERT INTO slides (title, description, image) VALUES (?, ?, ?)", Title, Description, Image)
	if err != nil {
		return
	}

	rows, err := db.Query("SELECT id FROM slides WHERE title=? AND description=? AND image=? ORDER BY id DESC", Title, Description, Image)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	err = rows.Scan(&id)
	if err != nil {
		return
	}

	err = UpdateSlides()
	return
}

// EditSlide updates a slide.
func EditSlide(ID int, Title, Description, Image string) (oldImage string, err error) {
	rows, err := db.Query("SELECT image FROM slides WHERE id=?", ID)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	err = rows.Scan(&oldImage)
	if err != nil {
		return
	}

	_, err = db.Exec("UPDATE slides SET title=?, description=?, image=? WHERE id=?", Title, Description, Image, ID)
	if err != nil {
		return
	}

	err = UpdateSlides()
	return
}

// EditSlideNoFile updates a slide without changing the file location.
func EditSlideNoFile(ID int, Title, Description string) (err error) {
	_, err = db.Exec("UPDATE slides SET title=?, description=? WHERE id=?", Title, Description, ID)
	if err != nil {
		return
	}

	err = UpdateSlides()
	return
}

// DeleteSlide deletes a slide.
func DeleteSlide(ID int) (image string, err error) {
	rows, err := db.Query("SELECT image FROM slides WHERE id=?", ID)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	err = rows.Scan(&image)
	if err != nil {
		return
	}

	_, err = db.Exec("DELETE FROM slides WHERE id=?", ID)
	if err != nil {
		return
	}

	err = UpdateSlides()
	return
}

// EditMenuItem updates a menu item.
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

	rows, err := db.Query("SELECT id FROM menu WHERE name=? AND description=? AND price=? ORDER BY id DESC", Name, Description, Price)
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

// EditUser updates a user.
func EditUser(ID int, Email, Password, Fname, Lname string, Privileges int) (err error) {
	_, err = db.Exec("UPDATE users SET email=?, password=?, fname=?, lname=?, priv=? WHERE uuid=?", Email, Password, Fname, Lname, Privileges, ID)
	if err != nil {
		return
	}

	err = UpdateUsers()
	return
}

// EditUserNoPassword updates a user without changing the password.
func EditUserNoPassword(ID int, Email, Fname, Lname string, Privileges int) (err error) {
	_, err = db.Exec("UPDATE users SET email=?, fname=?, lname=?, priv=? WHERE uuid=?", Email, Fname, Lname, Privileges, ID)
	if err != nil {
		return
	}

	err = UpdateUsers()
	return
}

// NewUser creates a new user.
func NewUser(Email, Password, Fname, Lname string, Privileges int) (id int, err error) {
	_, err = db.Exec("INSERT INTO users (email, password, fname, lname, priv) VALUES (?, ?, ?, ?, ?)", Email, Password, Fname, Lname, Privileges)
	if err != nil {
		return
	}

	rows, err := db.Query("SELECT uuid FROM users WHERE email=? AND password=? AND fname=? AND lname=? AND priv=? ORDER BY uuid DESC", Email, Password, Fname, Lname, Privileges)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	err = rows.Scan(&id)
	if err != nil {
		return
	}

	err = UpdateUsers()
	return
}

// DeleteUser deletes a user.
func DeleteUser(ID int) (err error) {
	_, err = db.Exec("DELETE FROM users WHERE uuid=?", ID)
	if err != nil {
		return
	}

	err = UpdateUsers()
	return
}
