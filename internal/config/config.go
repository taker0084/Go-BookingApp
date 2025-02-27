package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
	"github.com/taker0084/Go-BookingApp/internal/models"
)

//さまざまなファイルでデータの受け渡しを行いたい(データベースのコネクトなど)場合、
//下のようにAppConfigの中に書くのが良い
type AppConfig struct {
	UseCache bool
	TemplateCache map[string]*template.Template
	InfoLog *log.Logger
	ErrorLog *log.Logger
	InProduction bool
	Session *scs.SessionManager
	MailChan chan models.MailData
}