package handler

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/VolticFroogo/TRHT-Webserver/db"
	"github.com/VolticFroogo/TRHT-Webserver/helpers"
	"github.com/VolticFroogo/TRHT-Webserver/middleware"
	"github.com/VolticFroogo/TRHT-Webserver/middleware/myJWT"
	"github.com/VolticFroogo/TRHT-Webserver/models"
	"github.com/go-recaptcha/recaptcha"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

var (
	captchaSecret = os.Getenv("CAPTCHA_SECRET") // Captcha secret changed since I doxed myself. :P
	captcha       = recaptcha.New(captchaSecret)
)

type loginData struct {
	Email, Password, Captcha string
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

type slideData struct {
	ID                 int
	Title, Description string
	CImage             bool // Note: cImage is short for changeImage
}

// Start the server by handling the web server.
func Start() {
	r := mux.NewRouter()
	r.StrictSlash(true)

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

	r.Handle("/admin/slide/new", negroni.New(
		negroni.HandlerFunc(middleware.Form),
		negroni.Wrap(http.HandlerFunc(slideNew)),
	))
	r.Handle("/admin/slide/update", negroni.New(
		negroni.HandlerFunc(middleware.Form),
		negroni.Wrap(http.HandlerFunc(slideUpdate)),
	))
	r.Handle("/admin/slide/delete", http.HandlerFunc(slideDelete))

	r.Handle("/admin/menu/new", http.HandlerFunc(menuNew))
	r.Handle("/admin/menu/update", http.HandlerFunc(menuUpdate))
	r.Handle("/admin/menu/delete", http.HandlerFunc(menuDelete))

	r.Handle("/admin/contact-us/delete", http.HandlerFunc(contactDelete))

	r.Handle("/admin/user/new", http.HandlerFunc(userNew))
	r.Handle("/admin/user/update", http.HandlerFunc(userUpdate))
	r.Handle("/admin/user/delete", http.HandlerFunc(userDelete))

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	log.Printf("Server started...")
	http.ListenAndServe(":84", r)
}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("handler/templates/index.html") // Parse the HTML page
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Template parsing error", err)
		return
	}

	variables := models.TemplateVariables{
		Slides: db.Slides,
		Menu:   db.Menu,
	}
	err = t.Execute(w, variables) // Execute temmplate with variables
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Template execution error", err)
	}
}

func contactUs(w http.ResponseWriter, r *http.Request) {
	var message contactUsData                       // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&message) // Decode response to struct.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if message.Captcha == "" {
		successResponse(false, w, r)
		return // There is no captcha response
	}
	captchaSuccess, err := captcha.Verify(message.Captcha, r.Header.Get("CF-Connecting-IP")) // Check the captcha
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Recaptcha error", err)
	}
	if !captchaSuccess {
		successResponse(false, w, r)
		return // Unsuccessful recaptcha
	}

	err = db.NewContactMessage(message.Name, message.Email, message.Message)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Adding message to DB error", err)
	}

	successResponse(true, w, r)
}

func admin(w http.ResponseWriter, r *http.Request) {
	uuidString := context.Get(r, "uuid").(string)
	uuid, err := strconv.Atoi(uuidString)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Error converting string to int", err)
	}

	user, err := db.GetUserFromID(uuid)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Error getting user from ID", err)
	}

	t, err := template.ParseFiles("handler/templates/admin.html") // Parse the HTML page
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Template parsing error", err)
		return
	}

	csrfSecret, err := r.Cookie("csrfSecret")
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Reading cookie error", err)
		return
	}

	variables := models.TemplateVariables{
		User:            user,
		CsrfSecret:      csrfSecret.Value,
		Slides:          db.Slides,
		Menu:            db.Menu,
		ContactMessages: db.ContactMessages,
		Users:           db.Users,
	}
	err = t.Execute(w, variables) // Execute temmplate with variables
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Template execution error", err)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	refreshTokenString, err := r.Cookie("refreshToken")
	if err != nil {
		helpers.ThrowErr(w, r, "Reading cookie error", err)
		return
	}

	myJWT.DeleteJTI(refreshTokenString.Value) // Remove their old Refresh Token.

	middleware.WriteNewAuth(w, r, "", "", "")

	middleware.RedirectToLogin(w, r)
}

func login(w http.ResponseWriter, r *http.Request) {
	var credentials loginData                           // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&credentials) // Decode response to struct.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if credentials.Captcha == "" {
		successResponse(false, w, r)
		return // There is no captcha response.
	}
	captchaSuccess, err := captcha.Verify(credentials.Captcha, r.Header.Get("CF-Connecting-IP")) // Check the captcha.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Recaptcha error", err)
	}
	if !captchaSuccess {
		successResponse(false, w, r)
		return // Unsuccessful captcha.
	}

	user, err := db.GetUserFromEmail(credentials.Email)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Getting user from DB error", err)
		return
	}

	valid := helpers.CheckPassword(credentials.Password, user.Password)

	if valid {
		authTokenString, refreshTokenString, csrfSecret, err := myJWT.CreateNewTokens(strconv.Itoa(user.UUID))
		if err != nil {
			successResponse(false, w, r)
			helpers.ThrowErr(w, r, "Creating tokens error", err)
			return
		}

		middleware.WriteNewAuth(w, r, authTokenString, refreshTokenString, csrfSecret)

		successResponse(true, w, r)
		return
	}

	successResponse(false, w, r)
}

func slideNew(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024) // 10MB

	data := slideData{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
	}

	file, handle, err := r.FormFile("imageFile")
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Decoding image error", err)
		return
	}
	defer file.Close()

	imageID, err := helpers.GenerateRandomString(32)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Generating imageID error", err)
		return
	}

	fileLocation := "/img/slide/" + imageID + filepath.Ext(handle.Filename)
	err = saveFile(w, file, fileLocation)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Saving image error", err)
		return
	}

	id, err := db.NewSlide(data.Title, data.Description, fileLocation)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Editing slide error", err)
		return
	}

	res := responseWithID{
		Success: true,
		ID:      id,
	}
	resEnc, err := json.Marshal(res) // Encode response into JSON.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON encoding error", err)
		return
	}
	w.Write(resEnc) // Write JSON data to response writer.
}

func slideUpdate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024) // 10MB

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Converting ID string to int error", err)
		return
	}

	cImage, err := strconv.ParseBool(r.FormValue("cImage"))
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Converting cImage string to bool error", err)
		return
	}

	data := slideData{
		ID:          id,
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		CImage:      cImage,
	}

	if data.CImage {
		file, handle, err := r.FormFile("imageFile")
		if err != nil {
			successResponse(false, w, r)
			helpers.ThrowErr(w, r, "Decoding image error", err)
			return
		}
		defer file.Close()

		imageID, err := helpers.GenerateRandomString(32)
		if err != nil {
			successResponse(false, w, r)
			helpers.ThrowErr(w, r, "Generating imageID error", err)
			return
		}

		fileLocation := "/img/slide/" + imageID + filepath.Ext(handle.Filename)
		err = saveFile(w, file, fileLocation)
		if err != nil {
			successResponse(false, w, r)
			helpers.ThrowErr(w, r, "Saving image error", err)
			return
		}

		oldFileLocation, err := db.EditSlide(data.ID, data.Title, data.Description, fileLocation)
		if err != nil {
			successResponse(false, w, r)
			helpers.ThrowErr(w, r, "Editing slide error", err)
			return
		}

		err = deleteFile(oldFileLocation)
		if err != nil {
			successResponse(false, w, r)
			helpers.ThrowErr(w, r, "Deleting oldImage error", err)
			return
		}
	} else {
		err := db.EditSlideNoFile(data.ID, data.Title, data.Description)
		if err != nil {
			successResponse(false, w, r)
			helpers.ThrowErr(w, r, "Editing slide error", err)
			return
		}
	}

	successResponse(true, w, r)
}

func slideDelete(w http.ResponseWriter, r *http.Request) {
	var data models.SlideEdit                    // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		return
	}

	image, err := db.DeleteSlide(data.ID)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Deleting menu item error", err)
		return
	}

	err = deleteFile(image)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Deleting image error", err)
		return
	}

	successResponse(true, w, r)
}

func deleteFile(fileLocation string) (err error) {
	err = os.Remove("./static" + fileLocation)
	return
}

func saveFile(w http.ResponseWriter, file multipart.File, saveLocation string) (err error) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	err = ioutil.WriteFile("./static"+saveLocation, data, 0666)
	return
}

func menuUpdate(w http.ResponseWriter, r *http.Request) {
	var data models.MenuItemEdit                 // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		successResponse(false, w, r)
		return
	}

	err = db.EditMenuItem(data.ID, data.Name, data.Description, data.Price)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Editing menu item error", err)
		return
	}

	successResponse(true, w, r)
}

func menuNew(w http.ResponseWriter, r *http.Request) {
	var data models.MenuItemEdit                 // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		successResponse(false, w, r)
		return
	}

	id, err := db.NewMenuItem(data.Name, data.Description, data.Price)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Creating menu item error", err)
		return
	}

	res := responseWithID{
		Success: true,
		ID:      id,
	}
	resEnc, err := json.Marshal(res) // Encode response into JSON.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON encoding error", err)
		return
	}
	w.Write(resEnc) // Write JSON data to response writer.
}

func menuDelete(w http.ResponseWriter, r *http.Request) {
	var data models.MenuItemEdit                 // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		successResponse(false, w, r)
		return
	}

	err = db.DeleteMenuItem(data.ID)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Deleting menu item error", err)
		return
	}

	successResponse(true, w, r)
}

func contactDelete(w http.ResponseWriter, r *http.Request) {
	var data models.ContactMessageEdit           // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		successResponse(false, w, r)
		return
	}

	err = db.DeleteContactMessage(data.ID)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Deleting contact message error", err)
		return
	}

	successResponse(true, w, r)
}

func userUpdate(w http.ResponseWriter, r *http.Request) {
	var data models.UserEdit                     // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		successResponse(false, w, r)
		return
	}

	uuidString := context.Get(r, "uuid").(string)
	uuid, err := strconv.Atoi(uuidString)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Error converting string to int", err)
	}

	user, err := db.GetUserFromID(uuid)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Error getting user from ID", err)
	}

	if user.Priv != models.PrivSuperAdmin {
		// User isn't a super admin.
		successResponse(false, w, r)
		return
	}

	if data.Password == "" {
		err = db.EditUserNoPassword(data.ID, data.Email, data.Fname, data.Lname, data.Privileges)
		if err != nil {
			successResponse(false, w, r)
			helpers.ThrowErr(w, r, "Editing user (no password) error", err)
			return
		}
	} else {
		password, err := helpers.HashPassword(data.Password)
		if err != nil {
			successResponse(false, w, r)
			helpers.ThrowErr(w, r, "Hashing password error", err)
			return
		}

		err = db.EditUser(data.ID, data.Email, password, data.Fname, data.Lname, data.Privileges)
		if err != nil {
			successResponse(false, w, r)
			helpers.ThrowErr(w, r, "Editing user error", err)
			return
		}
	}

	successResponse(true, w, r)
	if err != nil {
		helpers.ThrowErr(w, r, "JSON encoding error", err)
	}
}

func userNew(w http.ResponseWriter, r *http.Request) {
	var data models.UserEdit                     // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		successResponse(false, w, r)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		successResponse(false, w, r)
		return
	}

	uuidString := context.Get(r, "uuid").(string)
	uuid, err := strconv.Atoi(uuidString)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Error converting string to int", err)
	}

	user, err := db.GetUserFromID(uuid)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Error getting user from ID", err)
	}

	if user.Priv != models.PrivSuperAdmin {
		// User isn't a super admin.
		successResponse(false, w, r)
		return
	}

	password, err := helpers.HashPassword(data.Password)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Hashing password error", err)
		return
	}

	id, err := db.NewUser(data.Email, password, data.Fname, data.Lname, data.Privileges)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Creating user error", err)
		return
	}

	res := responseWithID{
		Success: true,
		ID:      id,
	}
	resEnc, err := json.Marshal(res) // Encode response into JSON.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON encoding error", err)
		return
	}
	w.Write(resEnc) // Write JSON data to response writer.
}

func userDelete(w http.ResponseWriter, r *http.Request) {
	var data models.UserEdit                     // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		successResponse(false, w, r)
		return
	}

	uuidString := context.Get(r, "uuid").(string)
	uuid, err := strconv.Atoi(uuidString)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Error converting string to int", err)
	}

	user, err := db.GetUserFromID(uuid)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Error getting user from ID", err)
	}

	if user.Priv != models.PrivSuperAdmin {
		// User isn't a super admin.
		successResponse(false, w, r)
		return
	}

	err = db.DeleteUser(data.ID)
	if err != nil {
		successResponse(false, w, r)
		helpers.ThrowErr(w, r, "Deleting user error", err)
		return
	}

	successResponse(true, w, r)
}

func successResponse(valid bool, w http.ResponseWriter, r *http.Request) {
	res := response{
		Success: valid,
	}
	resEnc, err := json.Marshal(res) // Encode response into JSON.
	if err != nil {
		helpers.ThrowErr(w, r, "Sending success response error: %v", err)
	}
	w.Write(resEnc) // Write JSON data to response writer.
	return
}
