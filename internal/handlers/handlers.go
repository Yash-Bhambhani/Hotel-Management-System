package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"hotel_management_system/internal/config"
	"hotel_management_system/internal/drivers"
	"hotel_management_system/internal/forms"
	"hotel_management_system/internal/models"
	"hotel_management_system/internal/renderers"
	"hotel_management_system/internal/repository"
	"hotel_management_system/internal/repository/dbrepo"
	"log"
	"net/http"
	"strconv"
	"time"
)

var Repo *Repository

// yaha pe instead of storing interface
// i could have also simply stored a new struct for postgres
// reference:https://chat.openai.com/share/8eb4c946-6754-426f-ad85-6a1127d9da7d
type Repository struct {
	DB  repository.DatabaseRepo
	App *config.AppConfig
}

// NewRepository function ka kaam bas repo ko initialise karna hai main mai taaki baad mai usme data bhar sake
func NewRepository(app *config.AppConfig, db *drivers.DB) *Repository {
	return &Repository{
		dbrepo.NewPostGresRepo(db.SQL, app),
		app}
}

// NewHandlers ka use karke jo memory allot ki thi usme ye data feed kar  diya
func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	//remoteIP := r.RemoteAddr
	//m.App.Session.Put(r.Context(), "remote_IP", remoteIP)
	renderers.RenderTemplateWithLayout(w, r, "home.page.tmpl", &models.TemplateData{})
	//fmt.Fprintln(w, "This is Home page")
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	//renderers.RenderTemplate(w, "about.page.tmpl")
	//stringmap := make(map[string]string)
	//stringmap["title"] = "About"
	//remoteIP := m.App.Session.GetString(r.Context(), "remote_IP")
	//stringmap["remote_IP"] = remoteIP
	renderers.RenderTemplateWithLayout(w, r, "about.page.tmpl", &models.TemplateData{
		//StringMap: stringmap,
	})
	//fmt.Fprintln(w, "This is About page")
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	renderers.RenderTemplateWithLayout(w, r, "contact.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	renderers.RenderTemplateWithLayout(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	// In the PostAvailability handler
	layout := "2006-01-02" // Matches the date format from JavaScript
	start := r.Form.Get("start_date")
	end := r.Form.Get("end_date")

	startDate, err := time.Parse(layout, start)
	if err != nil {
		fmt.Println("Error in parsing start date 1")
		log.Fatal(err)
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		fmt.Println("Error in parsing end date")
		log.Fatal(err)
	}

	allRooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		fmt.Println("error in searching availability for all rooms", err)
		return
	}

	if len(allRooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "NO rooms available")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	reservationData := models.ReservationData{StartDate: startDate, EndDate: endDate}
	m.App.Session.Put(r.Context(), "reservationData", reservationData)
	//http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
	data := make(map[string]interface{})
	data["allRooms"] = allRooms
	renderers.RenderTemplateWithLayout(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
	//w.Write([]byte(fmt.Sprintf("the start point is %s and the end point is %s", start, end)))
}

type AvailableJSON struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomId    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		fmt.Println("Error in parsing start date 2")
		log.Fatal(err)
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		fmt.Println("Error in parsing end date")
		log.Fatal(err)
	}
	roomId, err := strconv.Atoi(r.Form.Get("roomID"))
	if err != nil {
		fmt.Println("Error in parsing room id")
	}
	available, err := m.DB.ParticularRoomAvailabilityByDate(startDate, endDate, roomId)
	if err != nil {
		fmt.Println("Error in getting available rooms")
		return
	}
	var response = AvailableJSON{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomId:    strconv.Itoa(roomId),
	}
	w.Header().Set("Content-Type", "application/json")
	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Println("Error in marshalling response")
		return
	}
	fmt.Println(string(output))
	_, err = w.Write(output)
	if err != nil {
		return
	}
}

func (m *Repository) Room(w http.ResponseWriter, r *http.Request) {
	renderers.RenderTemplateWithLayout(w, r, "room.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Suite(w http.ResponseWriter, r *http.Request) {
	renderers.RenderTemplateWithLayout(w, r, "suite.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	emptyTemplate, ok := m.App.Session.Get(r.Context(), "reservationData").(models.ReservationData)
	if !ok {
		log.Println("Going directly to make-reservation")
		m.App.Session.Put(r.Context(), "error", "NO reservations found")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	sd := emptyTemplate.StartDate.Format("2006-01-02")
	ed := emptyTemplate.EndDate.Format("2006-01-02")

	stringMap := map[string]string{
		"start_date": sd,
		"end_date":   ed,
	}

	roomInfo, err := m.DB.GetRoomByID(emptyTemplate.RoomID)
	if err != nil {
		fmt.Println("Error in getting room info")
		return
	}
	emptyTemplate.Room = roomInfo

	data := make(map[string]interface{})
	data["reservationData"] = emptyTemplate
	m.App.Session.Put(r.Context(), "reservationData", emptyTemplate)

	renderers.RenderTemplateWithLayout(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.NewForm(nil),
		Data:      data,
		StringMap: stringMap,
	})
}
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	//var emptyTemplate models.ReservationData
	emptyTemplate, ok := m.App.Session.Get(r.Context(), "reservationData").(models.ReservationData)
	if !ok {
		log.Println("Going directly to make-reservation")
		m.App.Session.Put(r.Context(), "error", "NO reservations found")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
		return
	}
	//sd := r.Form.Get("start_date")
	//ed := r.Form.Get("end_date")
	startDate := emptyTemplate.StartDate
	endDate := emptyTemplate.EndDate

	//layout := "2006-01-02"
	//startDate, err := time.Parse(layout, sd)
	//if err != nil {
	//	fmt.Println("Error in parsing start date 3")
	//	log.Fatal(err)
	//}
	//endDate, err := time.Parse(layout, ed)
	//if err != nil {
	//	fmt.Println("Error in parsing end date")
	//	log.Fatal(err)
	//}
	//roomId, err := strconv.Atoi(r.Form.Get("roomID"))
	//if err != nil {
	//	fmt.Println("Error in parsing room id")
	//}
	reservationData := models.ReservationData{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    emptyTemplate.RoomID,
	}
	form := forms.NewForm(r.PostForm)

	form.RequirementChecking("first_name", "last_name", "email", "phone")
	form.MinLength("phone", 10, r)
	form.IsValidEmail("email")
	if !form.IsValid() {
		data := make(map[string]interface{})
		data["reservationData"] = reservationData
		renderers.RenderTemplateWithLayout(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}
	newReservationID, err := m.DB.InsertReservation(reservationData)
	if err != nil {
		fmt.Println("Error in inserting reservation data in DB")
		return
	}
	roomRestriction := models.RoomRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		RoomID:        emptyTemplate.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}
	err = m.DB.InsertRoomRestriction(roomRestriction)
	if err != nil {
		fmt.Println("Error in inserting roomRestriction data in DB")
		return
	}
	reservationData.Room = emptyTemplate.Room
	reservationData.Room.RoomName = emptyTemplate.Room.RoomName

	m.App.Session.Put(r.Context(), "reservationData", reservationData)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservationData, ok := m.App.Session.Get(r.Context(), "reservationData").(models.ReservationData)
	// ok ke andar check hoga ki key(reservationData) ka data type uske baad wale
	//paranthesis ke andar ke datatype se match hona chahiye
	if !ok {
		log.Println("Error in getting reservationData")
		m.App.Session.Put(r.Context(), "error", "NO reservations found")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	data := make(map[string]interface{})
	data["reservationData"] = reservationData
	//fmt.Println(reservationData.Room.RoomName)
	m.App.Session.Remove(r.Context(), "reservationData")

	renderers.RenderTemplateWithLayout(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data})
}

func (m *Repository) ChooseAvailableRoom(w http.ResponseWriter, r *http.Request) {
	roomId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		fmt.Println("Error in parsing room id ")
		return
	}
	reservationData, ok := m.App.Session.Get(r.Context(), "reservationData").(models.ReservationData)
	if !ok {
		log.Println("Error in getting reservationData")
		m.App.Session.Put(r.Context(), "error", "error in getting reservationData")
		http.Redirect(w, r, "/choose-room/{id}", http.StatusTemporaryRedirect)
		return
	}
	reservationData.RoomID = roomId
	m.App.Session.Put(r.Context(), "reservationData", reservationData)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	renderers.RenderTemplateWithLayout(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.NewForm(nil),
	})
}

func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		fmt.Println("Error in parsing form")
		log.Println(err)
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.NewForm(r.PostForm)

	form.RequirementChecking("email", "password")
	form.IsValidEmail("email")

	if !form.IsValid() {
		renderers.RenderTemplateWithLayout(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}
	id, _, err := m.DB.AuthUser(email, password)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid Credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}
	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged In")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	err := m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())
	if err != nil {
		log.Println("Error in destroying session")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	renderers.RenderTemplateWithLayout(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	renderers.RenderTemplateWithLayout(w, r, "admin-new-reservation.page.tmpl", &models.TemplateData{})
}

func (m *Repository) AdminReservationsAll(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		fmt.Println("Error in getting reservations for dashboard")
		return
	}

	data := make(map[string]interface{})
	data["reservationData"] = reservations

	renderers.RenderTemplateWithLayout(w, r, "admin-all-reservation.page.tmpl", &models.TemplateData{
		Data: data,
	})
}
func (m *Repository) AdminReservationCalender(w http.ResponseWriter, r *http.Request) {
	renderers.RenderTemplateWithLayout(w, r, "admin-reservation-calender.page.tmpl", &models.TemplateData{})
}

func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	//query ke form mai data hai parameter ke nahi to ye nahi lagenge
	//roomId, err := strconv.Atoi(chi.URLParam(r, "id"))
	//if err != nil {
	//	fmt.Println("Error in parsing room id ", err)
	//	return
	//}
	//sd := chi.URLParam(r, "s")
	//ed := chi.URLParam(r, "e")
	query := r.URL.Query()

	// Accessing specific parameters
	roomId, err := strconv.Atoi(query.Get("id"))
	if err != nil {
		fmt.Println("Error in parsing room id ", err)
		return
	}
	sd := query.Get("s")
	ed := query.Get("e")
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		fmt.Println("Error in parsing start date")
		log.Fatal(err)
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		fmt.Println("Error in parsing end date")
		log.Fatal(err)
	}
	room, err := m.DB.GetRoomByID(roomId)
	if err != nil {
		log.Fatal("Error in getting room name", err)
	}
	reservationData := models.ReservationData{StartDate: startDate, EndDate: endDate, RoomID: roomId, Room: room}
	m.App.Session.Put(r.Context(), "reservationData", reservationData)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
	//fmt.Fprintln(w, "hello world")
}
