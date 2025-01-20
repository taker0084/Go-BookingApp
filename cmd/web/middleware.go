package main

import (
	"net/http"

	"github.com/justinas/nosurf"
)

//NoSurf adds CSRF protection to all POST request
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
	//requestにsessionを含める
	return session.LoadAndSave(next)
}