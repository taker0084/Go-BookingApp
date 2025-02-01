package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/taker0084/Go-BookingApp/internal/config"
	"github.com/taker0084/Go-BookingApp/internal/driver"
	"github.com/taker0084/Go-BookingApp/internal/handlers"
	"github.com/taker0084/Go-BookingApp/internal/helpers"
	"github.com/taker0084/Go-BookingApp/internal/models"
	"github.com/taker0084/Go-BookingApp/internal/render"

	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

//main is the main application function
func main(){
	db,err := run()

	if err != nil{
		log.Fatal((err))
	}
	defer db.SQL.Close()

	defer close(app.MailChan)

	fmt.Println("Starting mail listener...")
	listenForMail()


	fmt.Printf("Starting application on port %s \n", portNumber)
	//_=http.ListenAndServe(portNumber, nil)

	srv := &http.Server{
		Addr: portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB,error){
	//what am I going to put in production
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	//read flags
	inProduction := flag.Bool("production",true,"Application is in production")
	useCache := flag.Bool("cache", true, "Use template cache")
	dbHost := flag.String("dbhost","localhost", "Database host")
	dbName := flag.String("dbname","", "Database name")
	dbUser := flag.String("dbuser","", "Database user")
	dbPass := flag.String("dbPass","", "Database password")
	dbPort := flag.String("dbport","5432", "Database port")
	dbSSL := flag.String("dbssl","disable", "Database ssl settings (disable, prefer, require)")

	flag.Parse()

	if *dbName == "" || *dbUser == ""{
		fmt.Println("Missing requires flags")
		os.Exit(1)
	}

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	//change this to true when in production
	app.InProduction = *inProduction

	infoLog = log.New(os.Stdout, "INFO\t",log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour //24 hours
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction //production set to true

	app.Session = session

	//connect to database
	log.Println("Connecting to database...")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)
	db,err := driver.ConnectSQL(connectionString)
	if err !=nil{
		log.Fatal("Cannot connect to database! Dying...")
	}
	log.Println("Connected to database!")

	tc,err :=render.CreateTemplateCache()
	if err != nil{
		log.Fatal("cannot create template cache")
		return nil,err
	}

	app.TemplateCache = tc
	app.UseCache = *useCache

	//make new instance of App
	repo := handlers.NewRepo(&app,db)
	//send AppConfig and DB instance back to handlers
	handlers.NewHandlers(repo)
	//send AppConfig instance back to the render
	render.NewRenderer(&app)
	//send AppConfig instance back to the handler
	helpers.NewHelpers(&app)

	// http.HandleFunc("/",handlers.Repo.Home)
	// http.HandleFunc("/about",handlers.Repo.About)
	return db,nil
}