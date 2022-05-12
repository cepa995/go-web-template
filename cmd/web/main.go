package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/cepa995/go-web-template/internal/config"
	"github.com/cepa995/go-web-template/internal/driver"
	"github.com/cepa995/go-web-template/internal/handlers"
	"github.com/cepa995/go-web-template/internal/helpers"
	"github.com/cepa995/go-web-template/internal/models"
	render "github.com/cepa995/go-web-template/internal/renderer"
)

var app config.AppConfig        // Application Configuration
var session *scs.SessionManager // Session Manager

func main() {
	db, portNumber, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	defer close(app.MailChan)
	listenForMail()

	app.InfoLog.Printf("Starting application on port %s", portNumber)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func run() (*driver.DB, string, error) {
	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime)
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	// Read Flags
	inProduction := flag.Bool("production", true, "Application is in production")
	useCache := flag.Bool("cache", false, "Use template cache")
	portNumber := flag.String("portNumber", ":8080", "")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPassword := flag.String("dbpassword", "", "Database password")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

	flag.StringVar(&app.SecretKey, "secret", "", "secret key for hashing email data")
	flag.StringVar(&app.FrontEnd, "frontend", "", "URL to front end")
	flag.StringVar(&app.SMTP.Host, "smtphost", "", "smtp host")
	flag.StringVar(&app.SMTP.Username, "smtpuser", "", "smtp user")
	flag.StringVar(&app.SMTP.Password, "smtppass", "", "smtp password")
	flag.IntVar(&app.SMTP.Port, "smtpport", 587, "smtp port")
	flag.Parse()

	if *dbName == "" || *dbPassword == "" || *dbUser == "" {
		app.ErrorLog.Println("Missing required flags")
		os.Exit(1)
	}

	app.InProduction = *inProduction

	// Step 1. Create User Session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	// Step 2. Connect to the database
	app.InfoLog.Println("Trying to Connect to PostgreSQL Database ")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPassword, *dbSSL)
	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		app.ErrorLog.Fatal("Cannot connect to PostgreSQL database")
	}
	app.InfoLog.Println("Successfully Connected to PostgreSQL Database")

	// 3. Save a session to the database
	session.Store = postgresstore.NewWithCleanupInterval(db.SQL, 30*time.Minute)
	app.Session = session

	// Step 3. Create Template Cache
	tc, err := render.CreateTemplateCache()
	if err != nil {
		app.ErrorLog.Fatal(fmt.Sprintf("Cannot create Template Cache due to - %v", err))
	}
	app.TemplateCache = tc
	app.UseCache = *useCache

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, *portNumber, nil

}
