package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"github.com/cepa995/go-web-template/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

//go:embed templates
var emailTemplateFS embed.FS

// listenForMail background function which listents for incoming models.MailData
func listenForMail() {
	go func() {
		for {
			msg := <-app.MailChan
			sendMail(msg)
		}
	}()
}

// sendMail instantiates mail server, construts an email based on models.MailData and sends it to specified user.
func sendMail(m models.MailData) {
	// Step 1. Create new SMTP client and configure it
	server := mail.NewSMTPClient()
	server.Host = app.SMTP.Host
	server.Port = app.SMTP.Port
	server.Username = app.SMTP.Username
	server.Password = app.SMTP.Password
	server.Encryption = mail.EncryptionTLS
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	// Step 2. Construct mail body in the format of HTML.
	templateToRender := fmt.Sprintf("templates/%s.html.gohtml", m.TemplateName)
	t, err := template.New("email-html").ParseFS(emailTemplateFS, templateToRender)
	if err != nil {
		app.ErrorLog.Println(err)
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", m.Data); err != nil {
		app.ErrorLog.Println(err)
	}

	formattedMessage := tpl.String()
	// Step 3. Connect to SMTP server and create SMTP client
	smtpClient, err := server.Connect()
	if err != nil {
		app.ErrorLog.Println(err)
	}

	// Step 4. Construct empty email and populate it with content defined in previous steps
	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
	email.SetBody(mail.TextHTML, formattedMessage)

	// Step 5. Send an email
	err = email.Send(smtpClient)
	if err != nil {
		app.ErrorLog.Println(err)
	}

	fmt.Println("Successfully sent email!")
}
