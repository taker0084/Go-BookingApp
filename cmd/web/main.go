package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/taker0084/Go-BookingApp/internal/config"
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
	err := run()

	if err != nil{
		log.Fatal((err))
	}

	fmt.Printf("Starting application on port %s \n", portNumber)
	//_=http.ListenAndServe(portNumber, nil)

	srv := &http.Server{
		Addr: portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() error{
	//what am I going to put in production
	gob.Register(models.Reservation{})
	//change this to true when in production
	app.InProduction = false

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

	tc,err :=render.CreateTemplateCache()
	if err != nil{
		log.Fatal("cannot create template cache")
		return err
	}

	app.TemplateCache = tc
	app.UseCache = false

	//make new instance of App
	repo := handlers.NewRepo(&app)
	//send AppConfig instance back to handlers
	handlers.NewHandlers(repo)

	//send AppConfig instance back to the render
	render.NewTemplates(&app)

	helpers.NewHelpers(&app)

	// http.HandleFunc("/",handlers.Repo.Home)
	// http.HandleFunc("/about",handlers.Repo.About)
	return nil
}