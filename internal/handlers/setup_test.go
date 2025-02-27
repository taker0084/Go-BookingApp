package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/justinas/nosurf"
	"github.com/taker0084/Go-BookingApp/internal/config"
	"github.com/taker0084/Go-BookingApp/internal/models"
	"github.com/taker0084/Go-BookingApp/internal/render"
)

var app config.AppConfig
var session *scs.SessionManager
var pathToTemplates = "./../../templates"
var functions = template.FuncMap{
	"humanDate": render.HumanDate,
	"formatDate": render.FormatDate,
	"iterate": render.Iterate,
}

func TestMain(m *testing.M){
	//what am I going to put in production
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})
	//change this to true when in production
	app.InProduction = false

	infoLog := log.New(os.Stdout, "INFO\t",log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour //24 hours
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction //production set to true

	app.Session = session

	mailChan := make(chan models.MailData)
	app.MailChan=mailChan
	defer close(mailChan)

	listenForMail()

	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache:", err)
	}

	app.TemplateCache = tc
	app.UseCache = true

	//make new instance of App
	repo := NewTestRepo(&app)
	//send AppConfig instance back to handlers
	NewHandlers(repo)

	//send AppConfig instance back to the render
	render.NewRenderer(&app)

	os.Exit(m.Run())
}

func listenForMail(){
	go func(){
		for{
			_ = <- app.MailChan
		}
	}()
}

func getRoutes() http.Handler{
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	//mux.Use(NoSurf)
	mux.Use(SessionLoad)


	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/generals-quarters", Repo.Generals)
	mux.Get("/majors-suite", Repo.Majors)
	mux.Get("/search-availability", Repo.Availability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Post("/search-availability-json", Repo.AvailabilityJSON)

	mux.Get("/contact", Repo.Contact)

	mux.Get("/make-reservation", Repo.Reservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostShowLogin)
	mux.Get("/user/logout", Repo.Logout)

	mux.Get("/admin/dashboard", Repo.AdminDashboard)
		mux.Get("/admin/reservations-new", Repo.AdminNewReservations)
		mux.Get("/admin/reservations-all", Repo.AdminAllReservations)
		mux.Get("/admin/reservations-calendar", Repo.AdminReservationCalendar)
		mux.Post("/admin/reservations-calendar", Repo.AdminPostReservationsCalendar)

		mux.Get("/admin/process-reservation/{src}/{id}/do",Repo.AdminProcessReservation)
		mux.Get("/admin/delete-reservation/{src}/{id}/do",Repo.AdminDeleteReservation)

		mux.Get("/admin/reservations/{src}/{id}/show", Repo.AdminShowReservation)
		mux.Post("/admin/reservations/{src}/{id}", Repo.AdminPostShowReservation)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}

func NoSurf(next http.Handler) http.Handler{
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: app.InProduction, //production set to true
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

//SessionLoad loads and saves the sessions on every request
func SessionLoad(next http.Handler) http.Handler{
	return session.LoadAndSave(next)
}

func CreateTestTemplateCache() (map[string]*template.Template,error){
	myCache := map[string]*template.Template{}

	//get all of the files named *.page.tmpl from ./templates
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl",pathToTemplates))
	if err != nil{
		return myCache, err
	}

	//range through all files ending with *.page.tmpl
	for _, page := range pages{
		//get file name
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil{
			return myCache, err
		}
		//get all layout files ending with *.layout.tmpl
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl",pathToTemplates))
		if err != nil{
			return myCache, err
		}
		//if find layout files, associate with templates and layouts
		if len(matches) > 0{
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl",pathToTemplates))
			if err != nil{
				return myCache, err
			}
		}
		//add template to Templates Set
		myCache[name] = ts
	}
	return myCache, nil
}