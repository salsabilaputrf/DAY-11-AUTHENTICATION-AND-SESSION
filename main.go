package main

import (
	"authentication-session/connection"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	// "log"
	"net/http"
	//"strconv"
	"text/template"
	//"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type MetaData struct{
	Title string
	IsLogin bool
	UserName string
	FlashData string
}

var Data = MetaData{
}

type dataProject struct {
	Id           int
	ProjectName  string
	StartDate    time.Time
	EndDate      time.Time
	Description  string
	Technologies []string
	Duration     string
}

type User struct {
    Id       int
    Name     string
    Email    string
    Password string
}



func main() {
	route := mux.NewRouter()

	connection.DatabaseConnect()

	// static folder
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	// routing
	route.HandleFunc("/", home).Methods("GET").Name("home")
	route.HandleFunc("/contact", contactMe).Methods("GET")
	route.HandleFunc("/addProject", addProject).Methods("GET")
	route.HandleFunc("/addProject", addProjectInput).Methods("POST")
	route.HandleFunc("/detailProject/{id}", detailProject).Methods("GET")
	route.HandleFunc("/deleteProject/{id}", deleteProject).Methods("GET")
	route.HandleFunc("/editProject/{id}", editProject).Methods("GET")
	route.HandleFunc("/editProjectInput/{id}", editProjectInput).Methods("POST")

	route.HandleFunc("/register", formRegister).Methods("GET")
	route.HandleFunc("/register", register).Methods("POST")
	route.HandleFunc("/login", formLogin).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")
	route.HandleFunc("/logout", logout).Methods("GET")

	// port := 5000
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe("localhost:8080", route)
}

func selisih(start time.Time, end time.Time) string {

	distance := end.Sub(start)

	// Menghitung durasi
	var duration string
	year := int(distance.Hours() / (12 * 30 * 24))
	if year != 0 {
		duration = strconv.Itoa(year) + " tahun"
	} else {
		month := int(distance.Hours() / (30 * 24))
		if month != 0 {
			duration = strconv.Itoa(month) + " bulan"
		} else {
			week := int(distance.Hours() / (7 * 24))
			if week != 0 {
				duration = strconv.Itoa(week) + " minggu"
			} else {
				day := int(distance.Hours() / (24))
				if day != 0 {
					duration = strconv.Itoa(day) + " hari"
				}
			}
		}
	}

	return duration
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("view/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
    session, _ := store.Get(r, "SESSION_ID")

    if session.Values["IsLogin"] != true {
        Data.IsLogin = false
    } else {
        Data.IsLogin = session.Values["IsLogin"].(bool)
        Data.UserName = session.Values["Name"].(string)
    }
	

	var result []dataProject

	rows, err := connection.Conn.Query(context.Background(), "SELECT * FROM tb_project ")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for rows.Next() {
		var each = dataProject{}

		var err = rows.Scan(&each.Id, &each.ProjectName, &each.StartDate, &each.EndDate, &each.Description, &each.Technologies)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		each.Duration = selisih(each.StartDate, each.EndDate)
		result = append(result, each)
	}

	respData := map[string]interface{}{
		"Data":     Data,
		"Projects": result,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, respData)
}

func contactMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("view/contact-form.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	// var store = sessions.NewCookieStore([]byte("SESSION_ID"))
    // session, _ := store.Get(r, "SESSION_ID")

	// if session.Values["IsLogin"] != true {
    //     Data.IsLogin = false
    // } else {
    //     Data.IsLogin = session.Values["IsLogin"].(bool)
    //     Data.UserName = session.Values["Name"].(string)
    // }

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

func addProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("view/addProject.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
    session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
        Data.IsLogin = false
    } else {
        Data.IsLogin = session.Values["IsLogin"].(bool)
        Data.UserName = session.Values["Name"].(string)
    }

	

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

func addProjectInput(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	projectName := r.PostForm.Get("projectName")
	startDate := r.PostForm.Get("startDate")
	endDate := r.PostForm.Get("endDate")
	desc := r.PostForm.Get("desc")
	tech := r.Form["technologi"]

	// Parsing string to time
	// Start Date
	startDateTime, _ := time.Parse("2006-01-02", startDate)

	// End Date
	endDateTime, _ := time.Parse("2006-01-02", endDate)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_project(project_name, start_date, end_date, description, technologies) VALUES ($1, $2, $3, $4, $5)", projectName, startDateTime, endDateTime, desc, tech)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func detailProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("view/detail-project.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	ID, _ := strconv.Atoi(mux.Vars(r)["id"])

	ProjectDetail := dataProject{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, project_name, start_date, end_date, description, technologies FROM tb_project WHERE id=$1", ID).Scan(
		&ProjectDetail.Id, &ProjectDetail.ProjectName, &ProjectDetail.StartDate, &ProjectDetail.EndDate, &ProjectDetail.Description, &ProjectDetail.Technologies)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	ProjectDetail.Duration = selisih(ProjectDetail.StartDate, ProjectDetail.EndDate)

	respDataDetail := map[string]interface{}{
		"Data":          Data,
		"ProjectDetail": ProjectDetail,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, respDataDetail)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_project WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func editProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("view/editProject.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	ID, _ := strconv.Atoi(mux.Vars(r)["id"])

	ProjectDetail := dataProject{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_project WHERE id=$1", ID).Scan(
		&ProjectDetail.Id, &ProjectDetail.ProjectName, &ProjectDetail.StartDate, &ProjectDetail.EndDate, &ProjectDetail.Description, &ProjectDetail.Technologies)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	respData := map[string]interface{}{
		"Data":          Data,
		"ProjectDetail": ProjectDetail,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, respData)
}

func editProjectInput(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	projectName := r.PostForm.Get("projectName")
	startDate := r.PostForm.Get("startDate")
	endDate := r.PostForm.Get("endDate")
	desc := r.PostForm.Get("desc")
	tech := r.Form["technologi"]

	// Parsing string to time
	// Start Date
	startDateTime, _ := time.Parse("2006-01-02", startDate)

	// End Date
	endDateTime, _ := time.Parse("2006-01-02", endDate)

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_project SET project_name = $1, start_date = $2, end_date = $3, description = $4, technologies = $5 WHERE id=$6", projectName, startDateTime, endDateTime, desc, tech, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func formRegister(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("view/register.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	respData := map[string]interface{}{
		"Data":     Data,
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, respData)
}

func formLogin(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("view/login.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	respData := map[string]interface{}{
		"Data":     Data,
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, respData)
}

func register(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
    if err != nil {
        log.Fatal(err)
    }

    name := r.PostForm.Get("Name")
    email := r.PostForm.Get("Email")
	password := r.PostForm.Get("Password")

    passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(name, email, password) VALUES ($1, $2, $3)", name, email, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
}

func login(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
    if err != nil {
        log.Fatal(err)
    }

    email := r.PostForm.Get("Email")
	password := r.PostForm.Get("Password")

    user := User{}

    err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email=$1", email).Scan(
        &user.Id, &user.Name, &user.Email, &user.Password,
    )
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("message : " + err.Error()))
        return
    }

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("message : " + err.Error()))
        return
    }

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
    session, _ := store.Get(r, "SESSION_ID")

	session.Values["IsLogin"] = true
    session.Values["Name"] = user.Name
	session.Values["Id"] = user.Id
    session.Options.MaxAge = 10800 

    session.AddFlash("Login success", "message")
    session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request){
	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
    session, _ := store.Get(r, "SESSION_ID")

    session.Options.MaxAge = -1 

    session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}