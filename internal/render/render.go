package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/justinas/nosurf"
	"github.com/taker0084/Go-BookingApp/internal/config"
	"github.com/taker0084/Go-BookingApp/internal/models"
)
var functions = template.FuncMap{
	"humanDate": HumanDate,
	"formatDate": FormatDate,
	"iterate": Iterate,
}

var app *config.AppConfig
var pathToTemplates = "./templates"

func Iterate(count int)[]int{
	var i int
	var items []int
	for i=1; i<=count; i++{
		items = append(items, i)
	}
	return items
}
//NewTemplates set the config for the template package
func NewRenderer(a *config.AppConfig){
	app = a
}


//HumanDate returns time in YYYY-MM-DD format
func HumanDate(t time.Time) string{
	return t.Format("2006-01-02")
}

func FormatDate(t time.Time, f string)string{
	return t.Format(f)
}

//if needed, add some actions in this function
func AddDefaultData(td *models.TemplateData,r *http.Request) *models.TemplateData{
	//example
	//td.Flash="success"
	//td.CSRFToken = "**********"
	td.Flash=app.Session.PopString(r.Context(),"flash")
	td.Error=app.Session.PopString(r.Context(),"error")
	td.Warning=app.Session.PopString(r.Context(),"warning")
	td.CSRFToken = nosurf.Token(r)
	if app.Session.Exists(r.Context(), "user_id"){
		td.IsAuthenticated = 1
	}
	return td
}
//RenderTemplates renders templates using html/template
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error{
	var tc map[string]*template.Template
	if app.UseCache{
		//create a template cache
		tc = app.TemplateCache
	} else {
		tc,_ = CreateTemplateCache()
	}

	//get requested template from cache
	t,ok := tc[tmpl]
	if !ok{
		return errors.New("can't get template from cache")
	}

	td = AddDefaultData(td, r)

	buf := new(bytes.Buffer)
	err := t.Execute(buf,td)
	if err!=nil{
		log.Fatal(err)
	}

	//render the template
	_, err = buf.WriteTo(w)
	if err != nil{
		fmt.Println("Error writing template to browser", err)
		return err
	}
	return nil
}

func CreateTemplateCache() (map[string]*template.Template,error){
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

// var tc = make(map[string]*template.Template)

// func RenderTemplate(w http.ResponseWriter, t string){
// 	var tmpl *template.Template
// 	var err error

// 	//check to see if we already have the template in our cache
// 	_, inMap := tc[t]
// 	if !inMap{
// 		//need to create the template
// 		log.Println("creating template and adding to cache")
// 		err = createTemplateCache(t)
// 		if err != nil{
// 			log.Println(err)
// 		}
// 	}else{
// 		//we have the template in the cache
// 		log.Println("using cached template")
// 	}

// 	tmpl = tc[t]

// 	err = tmpl.Execute(w, nil)
// 	if err != nil{
// 			log.Println(err)
// 	}
// }

// func createTemplateCache(t string) error{
// 	templates := []string{
// 		fmt.Sprintf("./templates/%s", t),
// 		"./templates/base.layout.tmpl",
// 	}

// 	//parse the template
// 	//... → 配列を分解
// 	tmpl, err := template.ParseFiles(templates...)
// 	if err != nil{
// 		return err
// 	}

// 	//add template to cache (map)
// 	tc[t] = tmpl
// 	return nil
// }