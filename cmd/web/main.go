package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"hotel_management_system/internal/config"
	"hotel_management_system/internal/drivers"
	"hotel_management_system/internal/handlers"
	"hotel_management_system/internal/models"
	"hotel_management_system/internal/renderers"
	"log"
	"net/http"
	"time"
)

const portnumber = ":8000"

var app config.AppConfig
var session *scs.SessionManager

func main() {

	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	// session ke andar non primitive data types "put" karne se phle batna padta hai
	gob.Register(models.ReservationData{})
	gob.Register(models.Room{})
	gob.Register(models.User{})
	gob.Register(models.Restriction{})
	app.Session = session

	TempCache, err := renderers.CreateTemplateCache()
	if err != nil {
		fmt.Println("Error creating template cache")
		return
	}
	app.UseCache = false
	app.TemplateCache = TempCache

	//Connecting to Database
	fmt.Println("Connecting to database ...")
	db, err := drivers.ConnectSQL("host=localhost port=5432 dbname=dbms_lab user=yash password=123")
	if err != nil {
		fmt.Println("Error connecting to database in main.go")
	}
	defer func(SQL *sql.DB) {
		err := SQL.Close()
		if err != nil {
			fmt.Println("Error closing database in main.go")
		}
	}(db.SQL)

	//repo ke naam pe initialize/space allot kar diya
	repo := handlers.NewRepository(&app, db)
	// new handlers ka use karke data handlers ko de diya
	handlers.NewHandlers(repo)
	renderers.NewTemplates(&app)

	//http.HandleFunc("/", handlers.Repo.Home)
	//http.HandleFunc("/about", handlers.Repo.About)
	//fmt.Println("starting server on port " + portnumber)
	//err = http.ListenAndServe(portnumber, nil)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	srv := http.Server{
		Addr:    portnumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
