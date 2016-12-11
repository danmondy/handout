package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"strings"
	"encoding/base64"
	"github.com/gorilla/mux"
)

const DOCROOT = "public" //this is where the non compileable stuff goes - probably /var/www/handout/ or /etc/handout/public

//========ROUTES========
func main() {
	r := mux.NewRouter()

	//endpoints
	r.HandleFunc("/", BasicAuth(ListFilesHandler))
	r.HandleFunc("/edit", BasicAuth(EditFileHandler)).Methods("get")
	r.HandleFunc("/edit", BasicAuth(SaveFileHandler)).Methods("post")

	//static files
	fs := http.FileServer(http.Dir(DOCROOT))
	r.PathPrefix("/img").Handler(fs)
	r.PathPrefix("/css").Handler(fs)
	r.PathPrefix("/editormd").Handler(fs)
	r.PathPrefix("/js").Handler(fs)

	log.Println("Listening on localhost:7777")
	err := http.ListenAndServe("localhost:7777", r)
	if err != nil {
		log.Fatal(err)
	}
}

//----------------------

//=======ENDPOINTS/HANDLERS=======
//EditFileHandler
func ListFilesHandler(w http.ResponseWriter, r *http.Request, u User) {

	//List every file that this user is allowed to edit.
	//This needs to be rethought - this is temporary
	for _, d := range u.Directories {
		files, err := ioutil.ReadDir(d)
		if err != nil {
			log.Fatal(err)
		}
		w.Write([]byte("<h1>Files</h1>"))
		for _, f := range files {
			w.Write([]byte(fmt.Sprintf("<a href=\"/edit?filepath=%[1]s/%[2]s\">%[1]s/%[2]s </a><br>", d, f.Name())))
		}
	}
}

//EditFileHandler
//method: get
func EditFileHandler(w http.ResponseWriter, r *http.Request, u User) {
	vals := r.URL.Query()
	filepath := vals["filepath"][0]
	if u.CanEditFile(filepath) {
		bytes, err := ioutil.ReadFile(filepath)
		if err != nil {
			log.Fatal(err)
		}

		model := struct { //make a quick anonymous struct for the views use
			FilePath    string
			FileContent string
		}{filepath, string(bytes)}
		renderTemplate(w, "edit", model)
		return
	}

	w.Write([]byte("You are not allowed to edit this file."))
	//TODO: User templates or mustache or some tool to write html to the client instead of these bytes.
}

//SaveFileHandler: /edit
//method: post
//form
func SaveFileHandler(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	filepath := r.FormValue("filepath")
	data := r.FormValue("filecontent")

	fmt.Println(filepath)
	if user.CanEditFile(filepath) {
		err := ioutil.WriteFile(filepath, []byte(data), 0644)
		if err != nil {
			log.Fatal(err)
		}

		model := struct { //make a quick anonymous struct for the views use
			FilePath    string
			FileContent string
		}{filepath, data}
		renderTemplate(w, "edit", model)
		return
	}

	w.Write([]byte("You are not allowed to edit this file."))
	//TODO: User templates or mustache or some tool to write html to the client instead of these bytes.
}

//-------------------------------

//========TEMPORARY==========
func GetUsers() []User {
	users := []User{
		User{
			Name:        "Gabe",
			Pword:       "Hughes",
			Directories: []string{"userfiles"},
			Files:       []string{},
		},
	}
	return users
}

func GetUser(username string, password string) User {
	users := GetUsers()
	for _, u := range users {
		if (u.Name == username && u.Pword == password) {
			return u
		}
	}
	return User{
		Name:"invalid",
		Pword:"invalid",
		Directories: []string{},
		Files: []string{},
	}
}

//-------------------------------------------------------

//=======AUTHENTICATION=======

type AuthedHandlerFunc func(w http.ResponseWriter, r *http.Request, u User)

//Auth:
//This ensures that only authed users can access a handler (endpoint).
//It takes in a custom handler with an extra parameter (user) and fills that information in.
//It then converts it to a function that golang can associate with an endpoint (a HandlerFunc).
//Wrapping the handlers this way takes the authentication logic out of each individual endpoint.
func BasicAuth(h AuthedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		// TokenIndex:      0        1
		// Authorization: Basic QJSDOGJANBJZJ==
		authTokens := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(authTokens) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}
		usernamepassword, err := base64.StdEncoding.DecodeString(authTokens[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
		pair := strings.SplitN(string(usernamepassword), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}
		username := pair[0]
		password := pair[1]
		log.Println(username);
		log.Println(password);
		u := GetUser(username, password)
		if ValidUser(u) {
			h(w, r, u)
			return
		}
		http.Error(w, "Not authorized", 401)
		return
	}
}


//--------------------------

//==========TEMPLATES==========
//THIS IS NOT A GENERATED WEBSITE AT THE MOMENT

var templates *template.Template

func init() {
	FuncMap := BuildFuncMap()
	fmt.Println("Docroot:", DOCROOT)
	templates = template.Must(template.New("handout").Funcs(FuncMap).ParseGlob(fmt.Sprintf("%s/templates/*", DOCROOT)))
}

func BuildFuncMap() template.FuncMap {
	return template.FuncMap{
		"PrettyYear":  func(t time.Time) string { return t.Format("2006") },
		"PrettyMonth": func(m time.Time) string { return m.Month().String()[0:3] + "." },
		"Elipses":     func(s string) string { return fmt.Sprintf("%s...", []byte(s)[0:3]) },
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, model interface{}) error {
	err := templates.ExecuteTemplate(w, tmpl+".html", model)
	return err
}

//----------------------------
