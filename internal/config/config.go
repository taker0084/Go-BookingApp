package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
)

//さまざまなファイルでデータの受け渡しを行いたい(データベースのコネクトなど)場合、
//下のようにAppConfigの中に書くのが良い
type AppConfig struct {
	UseCache bool
	TemplateCache map[string]*template.Template
	InfoLog *log.Logger
	InProduction bool
	Session *scs.SessionManager
}