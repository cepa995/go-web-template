package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cepa995/go-web-template/internal/config"
	"github.com/cepa995/go-web-template/internal/driver"
	"github.com/cepa995/go-web-template/internal/encryption"
	"github.com/cepa995/go-web-template/internal/forms"
	"github.com/cepa995/go-web-template/internal/helpers"
	"github.com/cepa995/go-web-template/internal/models"
	render "github.com/cepa995/go-web-template/internal/render"
	"github.com/cepa995/go-web-template/internal/repository"
	"github.com/cepa995/go-web-template/internal/repository/dbrepo"
	"github.com/cepa995/go-web-template/internal/urlsigner"
	"golang.org/x/crypto/bcrypt"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository holds the application configuration and specifies which type of
// database are we actually using.
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewTestingRepo creates a new testing DB repository
func NewTestingRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// NewHandlers - sets repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

/*******************************************************************
                   BASIC RENDERING HANDLERS
********************************************************************/

// Home handler - renders home page.
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.gohtml", &models.TemplateData{})
}

/*******************************************************************
                   AUTHENTICATION HANDLERS
********************************************************************/

func (m *Repository) ShowAuth(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "auth.page.gohtml", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// SignOut handles user signing out
func (m *Repository) SignOut(w http.ResponseWriter, r *http.Request) {
	// Step 1. Destroy current user Session
	_ = m.App.Session.Destroy(r.Context())
	// Step 2. Renew Session token
	_ = m.App.Session.RenewToken(r.Context())
	// Step 3. Redirect to login page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// PostSignIn handles logging the user in.
func (m *Repository) PostSignIn(w http.ResponseWriter, r *http.Request) {
	// Prevetns session fixation attack. Every session thats stored anywhere in application has
	// associated a certain token with it. When doing login/logout its good practice to renew token.
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Could not parse the form")
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(w, r, "auth.page.gohtml", &models.TemplateData{
			Form: form,
		})
		return
	}

	email := form.Get("email")
	password := form.Get("password")

	// Step 1. Authenticate th user; get user by email and compare hashed password with password user provided
	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid Login credentials")
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	user, err := m.DB.GetUserByID(id)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", fmt.Sprintf("Could not get %s from the database", email))
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	// Step 4. Log in the user by storing userID in the session
	m.App.Session.Put(r.Context(), "user_id", user.ID)
	m.App.Session.Put(r.Context(), "access_level", user.AccessLevel)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// PostSignUp handler - renders sign in page
func (m *Repository) PostSignUp(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Could not parse the form")
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
	}

	form := forms.New(r.PostForm)
	form.IsEmail("email")
	form.Required(
		"firstName",
		"lastName",
		"email",
	)
	form.MinLength("firstName", 3)
	form.MinLength("lastName", 3)

	if !form.Valid() {
		var message string
		if form.Errors.Get("email") != "" {
			message = "Make sure email is properly formated."
		} else if form.Errors.Get("firstName") != "" || form.Errors.Get("lastName") != "" {
			message = "Make sure each field is at least 8 characters long."
		}

		resp := jsonResponse{
			OK:      false,
			Message: message,
		}

		out, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	firstName := form.Get("firstName")
	lastName := form.Get("lastName")

	email := form.Get("email")
	_, err = m.DB.GetUserByEmail(email)
	if err == nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Email address already exists!",
		}
		helpers.WriteJSON(w, http.StatusBadRequest, resp)
		return
	}

	link := fmt.Sprintf("%s/activate-account?firstName=%s&lastName=%s&email=%s", m.App.FrontEnd, firstName, lastName, email)

	sign := urlsigner.Signer{
		Secret: []byte(m.App.SecretKey),
	}
	signedLink := sign.GenerateTokenFromString(link)
	var data struct {
		Link string
	}
	data.Link = signedLink
	msg := models.MailData{
		To:           email,
		From:         "admin@muscle-factory.pro",
		Subject:      "Activate Account",
		TemplateName: "activate-account",
		Data:         data,
	}

	m.App.MailChan <- msg

	resp := jsonResponse{
		OK:      true,
		Message: "Success!",
	}
	helpers.WriteJSON(w, http.StatusOK, resp)
}

// ForgotPassword handles rendering forgot password page.
func (m *Repository) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "auth-forgot-password.page.gohtml", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// SendPasswordResetEmail handles sending link for reseting password via email
func (m *Repository) SendPasswordResetEmail(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Could not parse the form")
		http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
	}

	form := forms.New(r.PostForm)
	form.Required("email")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, r, "auth-forgot-password.page.gohtml", &models.TemplateData{
			Form: form,
		})
		return
	}

	email := form.Get("email")

	// Verify that User with specified email exists
	_, err = m.DB.GetUserByEmail(email)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", fmt.Sprintf("User with %s email does not exist", email))
		http.Redirect(w, r, "/forgot-password", http.StatusTemporaryRedirect)
		return
	}

	link := fmt.Sprintf("%s/reset-password?email=%s", m.App.FrontEnd, email)
	sign := urlsigner.Signer{
		Secret: []byte(m.App.SecretKey),
	}

	signedLink := sign.GenerateTokenFromString(link)

	var data struct {
		Link string
	}
	data.Link = signedLink
	msg := models.MailData{
		To:           email,
		From:         "admin@muscle-factory.pro",
		Subject:      "Password Reset Request",
		TemplateName: "password-reset",
		Data:         data,
	}

	m.App.MailChan <- msg
}

// ShowResetPassword handles rendering page for entering new password after clicking on reset link
func (m *Repository) ShowResetPassword(w http.ResponseWriter, r *http.Request) {
	theURL := r.RequestURI
	testURL := fmt.Sprintf("%s%s", m.App.FrontEnd, theURL)
	signer := urlsigner.Signer{
		Secret: []byte(m.App.SecretKey),
	}

	valid := signer.VerifyToken(testURL)
	if !valid {
		m.App.Session.Put(r.Context(), "error", "Invalid URL - tampering detected")
		http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
		return
	}

	// Step 2. Make sure password token has not expired
	expired := signer.IsExpired(testURL, 60)
	if expired {
		m.App.Session.Put(r.Context(), "error", "Link has expired")
		http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
		return
	}

	encryptor := encryption.Encryption{
		Key: []byte(m.App.SecretKey),
	}

	email := r.URL.Query().Get("email")
	encryptedEmail, err := encryptor.Encrypt(email)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Could not encrypt email")
		http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "email", encryptedEmail)
	render.Template(w, r, "auth-reset-password.page.gohtml", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// ResetPassword handles updating user password
func (m *Repository) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// Step 1. Parse the Form
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Could not parse the form")
		http.Redirect(w, r, "/reset-password", http.StatusSeeOther)
	}

	form := forms.New(r.PostForm)
	form.Required("password", "verify-password")

	if !form.Valid() {
		render.Template(w, r, "auth-reset-password.page.gohtml", &models.TemplateData{
			Form: form,
		})
		return
	}

	// Step 2. Make sure user has re-entered his password correctly
	newPassword := form.Get("password")
	verifyPassword := form.Get("verify-password")
	if newPassword != verifyPassword {
		render.Template(w, r, "auth-reset-password.page.gohtml", &models.TemplateData{
			Form: form,
		})
		return
	}

	// Step 3. Get user by email that has been stored in the session
	if !m.App.Session.Exists(r.Context(), "email") {
		render.Template(w, r, "auth-reset-password.page.gohtml", &models.TemplateData{
			Form: form,
		})
		return
	}
	encryptedEmail, ok := m.App.Session.Get(r.Context(), "email").(string)
	if !ok {
		m.App.ErrorLog.Println(fmt.Sprintf("could not convert interface - %v to string", m.App.Session.Get(r.Context(), "email")))
		helpers.ServerError(w, err)
		return
	}

	encryptor := encryption.Encryption{
		Key: []byte(m.App.SecretKey),
	}

	email, err := encryptor.Decrypt(encryptedEmail)
	if err != nil {
		m.App.ErrorLog.Println("could not decrypt email address")
		helpers.ServerError(w, err)
		return
	}

	user, err := m.DB.GetUserByEmail(email)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Step 4. Generate new password hash
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Step 5. Update the user password
	err = m.DB.UpdatePasswordForUser(user, string(newHash))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

// ShowActivateUserAccount handles rendering page when user visits link sent via his email for activating has account
func (m *Repository) ShowActivateUserAccount(w http.ResponseWriter, r *http.Request) {
	theURL := r.RequestURI

	testURL := fmt.Sprintf("%s%s", m.App.FrontEnd, theURL)
	testURL = strings.ReplaceAll(testURL, "amp;", "")

	signer := urlsigner.Signer{
		Secret: []byte(m.App.SecretKey),
	}

	// Step 1. Verify URL token
	valid := signer.VerifyToken(testURL)
	if !valid {
		m.App.Session.Put(r.Context(), "error", "Invalid URL - tampering detected")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Step 2. Make sure password token has not expired
	expired := signer.IsExpired(testURL, 60)
	if expired {
		m.App.Session.Put(r.Context(), "error", "Link has expired")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	encryptor := encryption.Encryption{
		Key: []byte(m.App.SecretKey),
	}

	firstName := r.URL.Query().Get("firstName")
	lastName := r.URL.Query().Get("lastName")
	email := r.URL.Query().Get("email")
	encryptedEmail, err := encryptor.Encrypt(email)
	if err != nil {
		m.App.ErrorLog.Println("could not decrypt email address")
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "firstName", firstName)
	m.App.Session.Put(r.Context(), "lastName", lastName)
	m.App.Session.Put(r.Context(), "email", encryptedEmail)
	render.Template(w, r, "auth-activate-account.page.gohtml", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// ActivateUserAccount handles inserting new user to the database
func (m *Repository) ActivateUserAccount(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Could not parse the form")
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("password")
	form.MinLength("password", 3)

	if !form.Valid() {
		resp := jsonResponse{
			OK:      false,
			Message: "Make sure your password is at least 3 characters long!",
		}
		helpers.WriteJSON(w, http.StatusBadRequest, resp)
		return
	}

	password := form.Get("password")
	firstName, ok := m.App.Session.Get(r.Context(), "firstName").(string)
	if !ok {
		m.App.ErrorLog.Println(fmt.Sprintf("could not convert interface - %v to string", m.App.Session.Get(r.Context(), "firstName")))
		return
	}
	lastName, ok := m.App.Session.Get(r.Context(), "lastName").(string)
	if !ok {
		m.App.ErrorLog.Println(fmt.Sprintf("could not convert interface - %v to string", m.App.Session.Get(r.Context(), "lastName")))
		return
	}
	encryptedEmail, ok := m.App.Session.Get(r.Context(), "email").(string)
	if !ok {
		m.App.ErrorLog.Println(fmt.Sprintf("could not convert interface - %v to string", m.App.Session.Get(r.Context(), "email")))
		helpers.ServerError(w, err)
		return
	}

	encryptor := encryption.Encryption{
		Key: []byte(m.App.SecretKey),
	}

	email, err := encryptor.Decrypt(encryptedEmail)
	if err != nil {
		m.App.ErrorLog.Println("could not decrypt email address")
		helpers.ServerError(w, err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		m.App.ErrorLog.Println("could not hash user password")
		helpers.ServerError(w, err)
		return
	}

	user := models.User{
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
		Password:    string(hashedPassword),
		AccessLevel: 1,
	}

	_, err = m.DB.InsertUser(user)
	if err != nil {
		m.App.ErrorLog.Println("could not insert user into the database")
		helpers.ServerError(w, err)
		return
	}

	resp := jsonResponse{
		OK:      true,
		Message: "Successfully registered user!",
	}

	helpers.WriteJSON(w, http.StatusBadRequest, resp)
}
