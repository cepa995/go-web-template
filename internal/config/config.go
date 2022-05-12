package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
	"github.com/cepa995/go-web-template/internal/models"
)

// SMTP holds SMTP server configuration
type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
}

// AppConfig holds the application configuration
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProduction  bool
	Session       *scs.SessionManager
	MailChan      chan models.MailData
	SMTP          SMTP
	SecretKey     string
	FrontEnd      string
}
