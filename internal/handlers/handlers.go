package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/taker0084/Go-BookingApp/internal/config"
	"github.com/taker0084/Go-BookingApp/internal/driver"
	"github.com/taker0084/Go-BookingApp/internal/forms"
	"github.com/taker0084/Go-BookingApp/internal/helpers"
	"github.com/taker0084/Go-BookingApp/internal/models"
	"github.com/taker0084/Go-BookingApp/internal/render"
	"github.com/taker0084/Go-BookingApp/internal/repository"
	"github.com/taker0084/Go-BookingApp/internal/repository/dbrepo"
)

//Repo the repository uses by the handlers
var Repo *Repository

//Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository{
	return &Repository{
		App: a,
		DB: dbrepo.NewPostgresRepo(db.SQL,a),
	}
}

// NewRepo creates a new repository
func NewTestRepo(a *config.AppConfig) *Repository{
	return &Repository{
		App: a,
		DB: dbrepo.NewTestingRepo(a),
	}
}

//NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository){
	Repo = r
}

//Home is the home page handler
func (m *Repository)Home(w http.ResponseWriter, r *http.Request){
	// remoteIP := r.RemoteAddr
	// m.App.Session.Put(r.Context(), "remote_ip", remoteIP)
	render.Template(w, r, "home.page.tmpl",&models.TemplateData{})
}

//About is the about page handler
func (m *Repository)About(w http.ResponseWriter, r *http.Request){
	//perform some logic
	// stringMap := make(map[string]string)
	// stringMap["test"] = "Hello, again."


	// remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")

	// stringMap["remote_ip"] = remoteIP

	//send the data to the templates
	render.Template(w, r,  "about.page.tmpl", &models.TemplateData{})
}

//Reservation renders the make a reservation page and displays form
func (m *Repository)Reservation(w http.ResponseWriter, r *http.Request){
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok{
		m.App.Session.Put(r.Context(),"error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room ,err := m.DB.GetRoomByID(res.RoomID)
	if err != nil{
		m.App.Session.Put(r.Context(),"error", "can't find room!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data:= make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r,  "make-reservation.page.tmpl",&models.TemplateData{
		Form: forms.New(nil),
		Data: data,
		StringMap: stringMap,
	})
}

//Reservation renders the make a reservation page and displays form
func (m *Repository)PostReservation(w http.ResponseWriter, r *http.Request){
	reservation,ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok{
		m.App.Session.Put(r.Context(), "error", "cannot get the reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()

	if err != nil{
		m.App.Session.Put(r.Context(), "error", "can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}


	reservation.FirstName=r.Form.Get("first_name")
	reservation.LastName=r.Form.Get("last_name")
	reservation.Email=r.Form.Get("email")
	reservation.Phone=r.Form.Get("phone")

	StartDate := r.Form.Get("start_date")
	EndDate := r.Form.Get("end_date")

	//Formの検証
	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name",3)
	form.IsEmail("email")

	if !form.Valid(){
		data := make(map[string]interface{})
		stringMap := make(map[string]string)
		stringMap["start_date"] = StartDate
		stringMap["end_date"] = EndDate
		data["reservation"] = reservation
		render.Template(w, r,  "make-reservation.page.tmpl",&models.TemplateData{
			Form: form,
			Data: data,
			StringMap: stringMap,
		})
		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil{
		m.App.Session.Put(r.Context(),"error", "can't insert reservation into database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// m.App.Session.Put(r.Context(),"reservation", reservation)

	restriction := models.RoomRestriction{
		StartDate: reservation.StartDate,
		EndDate: reservation.EndDate,
		RoomID: reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil{
		m.App.Session.Put(r.Context(),"error", "can't insert room restriction")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//send notifications - first to guest
	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		Dear %s:<br>
		This is confirm your reservation from %s to %s.
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To: reservation.Email,
		From: "me@here.com",
		Subject: "Reservation Confirmation",
		Content: htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	htmlMessage = fmt.Sprintf(`
		<strong>Reservation Notification</strong><br>
		A reservation has been made for %s from %s to %s.
	`, reservation.Room.RoomName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg = models.MailData{
		To: "me@here.com",
		From: "me@here.com",
		Subject: "Reservation Notification",
		Content: htmlMessage,
	}

	m.App.MailChan <- msg

	//渡すデータをセッションに保存
	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

//Generals renders the room page
func (m *Repository)Generals(w http.ResponseWriter, r *http.Request){
	render.Template(w, r,  "generals.page.tmpl",&models.TemplateData{})
}

//Majors renders the room page
func (m *Repository)Majors(w http.ResponseWriter, r *http.Request){
	render.Template(w, r,  "majors.page.tmpl",&models.TemplateData{})
}

//Availability renders the search availability page
func (m *Repository)Availability(w http.ResponseWriter, r *http.Request){
	render.Template(w, r,  "search-availability.page.tmpl",&models.TemplateData{})
}
//PostAvailability renders the search-availability page
func (m *Repository)PostAvailability(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"
	startDate,err := time.Parse(layout,start)
	if err != nil{
		m.App.Session.Put(r.Context(), "error", "can't parse start date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate,err := time.Parse(layout, end)
	if err != nil{
		m.App.Session.Put(r.Context(), "error", "can't parse end date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil{
		m.App.Session.Put(r.Context(), "error", "can't get availability for rooms")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if len(rooms) == 0{
		//No availability
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate:  startDate,
		EndDate: endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct{
	OK bool `json:"ok"`
	Message string `json:"message"`
	RoomID string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate string `json:"end_date"`
}

//AvailabilityJSON handlers request for availability and send JSON response
func (m *Repository)AvailabilityJSON(w http.ResponseWriter, r *http.Request){
	//need to parse request body
	err := r.ParseForm()
	if err !=nil{
		//can't parse form, so return appropriate json
		resp := jsonResponse{
			OK: false,
			Message: "Internal server error",
		}

		out,_ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("content-Type", "application/json")
		w.Write(out)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	layout := "2006-01-02"
	startDate,_ := time.Parse(layout, sd)
	endDate,_ := time.Parse(layout, ed)

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err !=nil{
		//can't parse form, so return appropriate json
		resp := jsonResponse{
			OK: false,
			Message: "Error querying database",
		}

		out,_ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("content-Type", "application/json")
		w.Write(out)
		return
	}
	resp := jsonResponse{
		OK: available,
		Message: "",
		StartDate: sd,
		EndDate: ed,
		RoomID: strconv.Itoa(roomID),
	}

	// I removed the error check, since we handle all aspect of the json right here
	out,_ := json.MarshalIndent(resp, "", "     ")

	w.Header().Set("content-Type", "application/json")
	w.Write(out)
}

//Majors renders the contact page
func (m *Repository)Contact(w http.ResponseWriter, r *http.Request){
	render.Template(w, r,  "contact.page.tmpl",&models.TemplateData{})
}

//ReservationSummary displays the reservation summary page
func (m *Repository)ReservationSummary(w http.ResponseWriter, r *http.Request){
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok{
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w,r,"/",http.StatusTemporaryRedirect)
		return
	}
	//summaryが取得できたらreservationの情報を削除
	m.App.Session.Remove(r.Context(), "reservation")

	data:=make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
		StringMap: stringMap,
	})
}

//ChooseRoom displays list available rooms
func (m *Repository)ChooseRoom(w http.ResponseWriter, r *http.Request){
	// changed to this, so we can test it more easily
	// split the URL up by /, and grab the 3rd element
	exploded := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}


	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok{
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom takes URL parameters, builds a sessional variables, and takes user to make res screen
func (m *Repository)BookRoom(w http.ResponseWriter, r *http.Request){
	roomID,_ := strconv.Atoi(r.URL.Query().Get("id"))

	startDate,_ := time.Parse("2006-01-02", r.URL.Query().Get("s"))

	endDate,_ := time.Parse("2006-01-02", r.URL.Query().Get("e"))

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil{
		m.App.Session.Put(r.Context(), "error", "Can't get room from db!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var reservation models.Reservation
	reservation.RoomID = roomID
	reservation.Room = room
	reservation.StartDate = startDate
	reservation.EndDate = endDate

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

//ShowLogin shows the login screen
func (m *Repository)ShowLogin(w http.ResponseWriter, r *http.Request){
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

//PostShowLogin handles logging the user in
func (m *Repository)PostShowLogin(w http.ResponseWriter, r *http.Request){
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil{
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid(){
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil{
		m.App.Session.Put(r.Context(), "error", "invalid credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

//Logout logs user out
func (m *Repository)Logout(w http.ResponseWriter, r *http.Request){
	m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository)AdminDashboard(w http.ResponseWriter, r *http.Request){
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

//AdminNewReservations shows all reservations in admin tool
func (m *Repository)AdminAllReservations(w http.ResponseWriter, r *http.Request){
	reservations,err := m.DB.AllReservations()
	if err != nil{
		helpers.ServerError(w,err)
		return
	}
	data := make(map[string]interface{})
	data["reservations"] = reservations
	render.Template(w, r, "admin-all-reservation.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

//AdminNewReservations shows all new reservations in admin tool
func (m *Repository)AdminNewReservations(w http.ResponseWriter, r *http.Request){
	reservations,err := m.DB.AllNewReservations()
	if err != nil{
		helpers.ServerError(w,err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations
	render.Template(w, r, "admin-new-reservation.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

//AdminShowReservation shows the reservation in the admin tool
func (m *Repository)AdminShowReservation(w http.ResponseWriter, r *http.Request){
	exploded := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(exploded[4])
	if err != nil{
		helpers.ServerError(w,err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	stringMap["month"] = month
	stringMap["year"] = year

	//get reservation from database
	res, err := m.DB.GetReservationByID(id)
	if err != nil{
		helpers.ServerError(w,err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "admin-show-reservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data: data,
		Form: forms.New(nil),
	})
}

//AdminShowReservation shows the reservation in the admin tool
func (m *Repository)AdminPostShowReservation(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err != nil{
		helpers.ServerError(w,err)
		return
	}
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil{
		helpers.ServerError(w,err)
		return
	}

	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	//get reservation from database
	res, err := m.DB.GetReservationByID(id)
	if err != nil{
		helpers.ServerError(w,err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil{
		helpers.ServerError(w, err)
		return
	}

	month := r.Form.Get("month")
	year := r.Form.Get("year")

		m.App.Session.Put(r.Context(), "flash", "Changes saved")

	if year ==""{
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	}else{
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year,month), http.StatusSeeOther)
	}

}

//AdminReservationCalendar displays the reservation calender
func (m *Repository)AdminReservationCalendar(w http.ResponseWriter, r *http.Request){
	//assume that there is no month/year specified
	now := time.Now()
	//if there is a month/year specified
	if r.URL.Query().Get("y") != ""{
		year,_ := strconv.Atoi(r.URL.Query().Get("y"))
		month,_ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data:=make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	LastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = LastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	//get the first and last days of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms,err := m.DB.AllRooms()
	if err != nil{
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	for _,x := range rooms{
		//create maps
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		//d=日付,d.After(lastOfMonth)は次の日が月の最終日かどうか
		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1){
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
		}

		//get all the restrictions for the current room
		restrictions,err := m.DB.GetRestrictionsForRoomByDate(x.ID,firstOfMonth,lastOfMonth)
		if err != nil{
			helpers.ServerError(w, err)
			return
		}

		for _,y := range restrictions{
			if y.ReservationID > 0{
				//it's a reservation
				for d := y.StartDate; !d.After(y.EndDate); d = d.AddDate(0, 0, 1){
					reservationMap[d.Format("2006-01-2")] = y.ReservationID
				}
			}else{
				//it's a block
				blockMap[y.StartDate.Format("2006-01-2")] = y.ID
			}
		}

		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
	}

	render.Template(w, r, "admin-reservations-calender.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data: data,
		IntMap: intMap,
	})
}

//AdminProcessReservation marks a reservation as processed
func (m *Repository)AdminProcessReservation(w http.ResponseWriter, r *http.Request){
	id, _ := strconv.Atoi(chi.URLParam(r,"id"))
	src := chi.URLParam(r,"src")
	err := m.DB.UpdateProcessedForReservation(id, 1)
	if err != nil{
		log.Println(err)
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")
	if year ==""{
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	}else{
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

//AdminDeleteReservation marks a reservation as processed
func (m *Repository)AdminDeleteReservation(w http.ResponseWriter, r *http.Request){
	id, _ := strconv.Atoi(chi.URLParam(r,"id"))
	src := chi.URLParam(r,"src")
	_ = m.DB.DeleteReservation(id)

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "Reservation deleted")

	if year ==""{
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	}else{
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

//AdminPostReservationsCalendar handles post of reservation calendar
func (m *Repository)AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err != nil{
		helpers.ServerError(w, err)
		return
	}

	year,_ := strconv.Atoi(r.PostForm.Get("y"))
	month,_ := strconv.Atoi(r.PostForm.Get("m"))

	rooms,err := m.DB.AllRooms()
	if err != nil{
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	for _,x := range rooms{
		//get the block map from the session. Loop through entire map. if we have an entry in the map
		//that does not exist in our posted data, and it's not zero, we need to delete the restriction by id
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)

		for name, value := range curMap{
			//ok will be false if the value is not in the map
			if val, ok := curMap[name]; ok{
				//only pay attention to values > 0, and those who are not in the form
				// the rest are just placeholders for days without blocks
				if val > 0{
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name)){
						//delete the restriction by id
						err := m.DB.DeleteBlockByID(value)
						if err != nil{
							log.Println(err)
						}
					}
				}
			}

	// 		//see if there is a need to create a new block
	// 		if form.Has(fmt.Sprintf("add_block_%d_%s", x.ID, d)){
	// 			//insert a new block
	// 			_ = m.DB.InsertBlockForRoom(x.ID, date.Date(year, time.Month(month), 1))
	// 		}
		}
	}

	//now handle new blocks
	for name := range r.PostForm{
		if strings.HasPrefix(name, "add_block"){
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			date, _ := time.Parse("2006-01-2", exploded[3])
			//insert a new block
			err = m.DB.InsertBlockForRoom(roomID, date)
			if err != nil{
				log.Println(err)
			}
		}
	}

	m.App.Session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}