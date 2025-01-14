package render

import (
	"bytes"
	"html/template"
	"log"
	"myapp/pkg/config"
	"myapp/pkg/models"
	"net/http"
	"path/filepath"
)

var app *config.AppConfig

//NewTemplates set the config for the template package
func NewTemplates(a *config.AppConfig){
	app = a
}

//if needed, add some actions in this function
func AddDefaultData(td *models.TemplateData) *models.TemplateData{
	//example
	//td.Flash="success"
	//td.CSRFToken = "**********"
	return td
}
//RenderTemplates renders templates using html/template
func RenderTemplate(w http.ResponseWriter, tmpl string, td *models.TemplateData){
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
		log.Fatal("could not get template from template cache")
	}

	td = AddDefaultData(td)

	buf := new(bytes.Buffer)
	_= t.Execute(buf,td)

	//render the template
	_, err := buf.WriteTo(w)
	if err != nil{
		log.Println("Error writing template to browser", err)
		return
	}
}

func CreateTemplateCache() (map[string]*template.Template,error){
	myCache := map[string]*template.Template{}

	//get all of the files named *.page.tmpl from ./templates
	pages, err := filepath.Glob("./templates/*.page.tmpl")
	if err != nil{
		return myCache, err
	}

	//range through all files ending with *.page.tmpl
	for _, page := range pages{
		//get file name
		name := filepath.Base(page)
		ts, err := template.New(name).ParseFiles(page)
		if err != nil{
			return myCache, err
		}
		//get all layout files ending with *.layout.tmpl
		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err != nil{
			return myCache, err
		}
		//if find layout files, associate with templates and layouts
		if len(matches) > 0{
			ts, err = ts.ParseGlob("./templates/*.layout.tmpl")
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