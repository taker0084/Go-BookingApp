package main

import (
	"fmt"
	"log"
	"myapp/pkg/config"
	"myapp/pkg/handlers"
	"myapp/pkg/render"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager

//main is the main application function
func main(){

	//change this to true when in production
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour //24 hours
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction //production set to true

	app.Session = session

	tc,err :=render.CreateTemplateCache()
	if err != nil{
		log.Fatal("cannot create template cache")
	}

	app.TemplateCache = tc
	app.UseCache = false

	//make new instance of App
	repo := handlers.NewRepo(&app)
	//send AppConfig instance back to handlers
	handlers.NewHandlers(repo)

	//send AppConfig instance back to the render
	render.NewTemplates(&app)

	// http.HandleFunc("/",handlers.Repo.Home)
	// http.HandleFunc("/about",handlers.Repo.About)

	fmt.Printf("Starting application on port %s \n", portNumber)
	//_=http.ListenAndServe(portNumber, nil)

	srv := &http.Server{
		Addr: portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}