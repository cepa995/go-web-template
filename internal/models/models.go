package models

import (
	"time"

	"github.com/cepa995/go-web-template/internal/forms"
)

const (
	ScopeAuthentication = "authentication"
)

// Users corresponds to users model
type User struct {
	ID          int64
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TemplateData contains data sent from handlers to templates
type TemplateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float32
	Data            map[string]interface{}
	CSRFToken       string
	Flash           string
	Warning         string
	Error           string
	Form            *forms.Form
	IsAuthenticated int
	API             string
	AccessLevel     int64
}

// MailData holds an email message
type MailData struct {
	To           string
	From         string
	Subject      string
	Data         interface{}
	TemplateName string
}
