package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/db"
	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/helpers"
	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/middleware"
	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/middleware/myJWT"
	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/models"
	"github.com/go-recaptcha/recaptcha"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

var (
	captchaSecret = os.Getenv("CAPTCHA_SECRET") // Captcha secret changed since I doxed myself. :P
	captcha       = recaptcha.New(captchaSecret)
)

type loginData struct {
	Email, Password string
}

type contactUsData struct {
	Name, Email, Message, Captcha string
}

type response struct {
	Success bool `json:"success"`
}

type responseWithID struct {
	Success bool `json:"success"`
	ID      int  `json:"id"`
}

// Start the server by handling the web server.
func Start() {
	log.Printf("Captcha: %v", captchaSecret)
	r := mux.NewRouter()

	r.Handle("/", http.HandlerFunc(index))
	r.Handle("/contact-us", http.HandlerFunc(contactUs))
	r.Handle("/loginajax", http.HandlerFunc(login))

	r.Handle("/logout", negroni.New(
		negroni.HandlerFunc(middleware.Form),
		negroni.Wrap(http.HandlerFunc(logout)),
	))

	r.Handle("/admin", negroni.New(
		negroni.HandlerFunc(middleware.Admin),
		negroni.Wrap(http.HandlerFunc(admin)),
	))

	r.Handle("/admin/menu", http.HandlerFunc(menuUpdate))
	r.Handle("/admin/menu/new", http.HandlerFunc(menuNew))
	r.Handle("/admin/menu/delete", http.HandlerFunc(menuDelete))

	r.Handle("/admin/contact-us/delete", http.HandlerFunc(contactDelete))

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	log.Printf("Server started...")
	http.ListenAndServe(":84", r)
}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("handler/templates/index.html") // Parse the HTML page
	if err != nil {
		helpers.ThrowErr(w, "Template parsing error", err)
		return
	}

	variables := models.TemplateVariables{
		Slides: db.Slides,
		Menu:   db.Menu,
	}
	err = t.Execute(w, variables) // Execute temmplate with variables
	if err != nil {
		helpers.ThrowErr(w, "Template execution error", err)
	}
}

func contactUs(w http.ResponseWriter, r *http.Request) {
	var message contactUsData                       // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&message) // Decode response to struct.
	if err != nil {
		helpers.ThrowErr(w, "JSON decoding error", err)
		return
	}

	if message.Captcha == "" {
		return // There is no captcha response
	}
	captchaSuccess, err := captcha.Verify(message.Captcha, r.Header.Get("CF-Connecting-IP")) // Check the captcha
	if err != nil {
		helpers.ThrowErr(w, "Recaptcha error", err)
	}
	if !captchaSuccess {
		return // Unsuccessful recaptcha
	}

	err = db.NewContactMessage(message.Name, message.Email, message.Message)
	if err != nil {
		helpers.ThrowErr(w, "Adding message to DB error", err)
	}

	err = successResponse(true, w)
	if err != nil {
		helpers.ThrowErr(w, "JSON encoding error", err)
	}
}

func admin(w http.ResponseWriter, r *http.Request) {
	authTokenString, err := r.Cookie("authToken")
	if err != nil {
		helpers.ThrowErr(w, "Reading cookie error", err)
		return
	}

	uuidString := myJWT.GetUUIDFromToken(authTokenString.Value)
	uuid, err := strconv.Atoi(uuidString)
	if err != nil {
		helpers.ThrowErr(w, "Error converting string to int", err)
	}

	user, err := db.GetUserFromID(uuid)
	if err != nil {
		helpers.ThrowErr(w, "Error getting user from ID", err)
	}

	t, err := template.ParseFiles("handler/templates/admin.html") // Parse the HTML page
	if err != nil {
		helpers.ThrowErr(w, "Template parsing error", err)
		return
	}

	csrfSecret, err := r.Cookie("csrfSecret")
	if err != nil {
		helpers.ThrowErr(w, "Reading cookie error", err)
		return
	}

	variables := models.TemplateVariables{
		User:            user,
		CsrfSecret:      csrfSecret.Value,
		Slides:          db.Slides,
		Menu:            db.Menu,
		ContactMessages: db.ContactMessages,
	}
	err = t.Execute(w, variables) // Execute temmplate with variables
	if err != nil {
		helpers.ThrowErr(w, "Template execution error", err)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	middleware.WriteNewAuth(w, r, "", "", "")

	middleware.RedirectToHome(w, r)
}

func login(w http.ResponseWriter, r *http.Request) {
	var credentials loginData                           // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&credentials) // Decode response to struct.
	if err != nil {
		helpers.ThrowErr(w, "JSON decoding error", err)
		return
	}

	user, err := db.GetUserFromEmail(credentials.Email)
	if err != nil {
		helpers.ThrowErr(w, "Getting user from DB error", err)
		return
	}

	valid := helpers.CheckPassword(credentials.Password, user.Password)

	if valid {
		authTokenString, refreshTokenString, csrfSecret, err := myJWT.CreateNewTokens(strconv.Itoa(user.UUID))
		if err != nil {
			helpers.ThrowErr(w, "Creating tokens error", err)
			return
		}

		middleware.WriteNewAuth(w, r, authTokenString, refreshTokenString, csrfSecret)
	}

	err = successResponse(true, w)
	if err != nil {
		helpers.ThrowErr(w, "JSON encoding error", err)
	}
}

func menuUpdate(w http.ResponseWriter, r *http.Request) {
	var data models.MenuItemEdit                 // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		helpers.ThrowErr(w, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		return
	}

	err = db.EditMenuItem(data.ID, data.Name, data.Description, data.Price)
	if err != nil {
		helpers.ThrowErr(w, "Editing menu item error", err)
		return
	}

	err = successResponse(true, w)
	if err != nil {
		helpers.ThrowErr(w, "JSON encoding error", err)
	}
}

func menuNew(w http.ResponseWriter, r *http.Request) {
	var data models.MenuItemEdit                 // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		helpers.ThrowErr(w, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		return
	}

	id, err := db.NewMenuItem(data.Name, data.Description, data.Price)
	if err != nil {
		helpers.ThrowErr(w, "Creating menu item error", err)
		return
	}

	res := responseWithID{
		Success: true,
		ID:      id,
	}
	resEnc, err := json.Marshal(res) // Encode response into JSON.
	if err != nil {
		helpers.ThrowErr(w, "JSON encoding error", err)
		return
	}
	w.Write(resEnc) // Write JSON data to response writer.
}

func menuDelete(w http.ResponseWriter, r *http.Request) {
	var data models.MenuItemEdit                 // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		helpers.ThrowErr(w, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		return
	}

	err = db.DeleteMenuItem(data.ID)
	if err != nil {
		helpers.ThrowErr(w, "Deleting menu item error", err)
		return
	}

	err = successResponse(true, w)
	if err != nil {
		helpers.ThrowErr(w, "JSON encoding error", err)
	}
}

func contactDelete(w http.ResponseWriter, r *http.Request) {
	var data models.ContactMessageEdit           // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		helpers.ThrowErr(w, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		return
	}

	err = db.DeleteContactMessage(data.ID)
	if err != nil {
		helpers.ThrowErr(w, "Deleting contact message error", err)
		return
	}

	err = successResponse(true, w)
	if err != nil {
		helpers.ThrowErr(w, "JSON encoding error", err)
	}
}

func successResponse(valid bool, w http.ResponseWriter) (err error) {
	res := response{
		Success: valid,
	}
	resEnc, err := json.Marshal(res) // Encode response into JSON.
	w.Write(resEnc)                  // Write JSON data to response writer.
	return
}
