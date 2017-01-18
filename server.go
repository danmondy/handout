package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"strconv"
	"net/http"
	"time"
	"strings"
	"encoding/base64"
	"os"
	"bufio"
	"crypto/sha1"
	"github.com/gorilla/mux"
)

const DOCROOT = "public" //this is where the non compileable stuff goes - probably /var/www/handout/ or /etc/handout/public
var usersFilename = ".handout.users"
var users []User

//========ROUTES========
func main() {
	port := 8080;
	portString := fmt.Sprint(port)
	if len(os.Args) >= 3 {
		log.Println("To many arguments");
		return
	}
	if len(os.Args) == 2 {
		s := os.Args[1]
		p, _ := strconv.ParseInt(s, 10, 64)
		if p >= 0 && p <= 10000 {
			port := p
			portString = fmt.Sprint(port)
		}
	}
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

	log.Println("Listening on localhost:"+portString)
	err := http.ListenAndServe("localhost:"+portString, r)
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
			log.Println(d)
			log.Println(files)
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

func GetUser(username string, password string) (User, bool) {
	for _, u := range users {
		if (u.Name == username && u.Pword == password) {
			return u, true
		}
	}
	return User{
		Name:"invalid",
		Pword:"invalid",
		Directories: []string{},
		Files: []string{},
	}, false
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
		passwordHash := sha1.Sum([]byte(password))
		passwordHashHex := fmt.Sprintf("%x", passwordHash)
		log.Println(username);
		log.Println(passwordHashHex);
		u, isUserValid := GetUser(username, passwordHashHex)
		if isUserValid {
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

func buildUsers() {
	// Build Users
	if _, err := os.Stat(usersFilename); os.IsNotExist(err) {
		fmt.Println("WARNING: No users file setup. A default users file was created")
		password := []byte("password")
		passwordHash := sha1.Sum(password)
		passwordHashHex := fmt.Sprintf("%x", passwordHash)
		admin := []byte("admin\t"+passwordHashHex+"\t\t\n")
		err := ioutil.WriteFile(usersFilename, admin, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	file, err := os.Open(usersFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	userCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 4 {
			userInfo := strings.Fields(line)
			var directories []string = []string{""}
			if len(userInfo) >= 3 && strings.Contains(userInfo[2], ":") {
				directories = strings.Split(userInfo[2], ":")
			} else {
				if len(userInfo) >= 3 {
					directories = []string{userInfo[2]}
				}
			}
			var files []string = []string{""}
			if len(userInfo) >= 4 && strings.Contains(userInfo[3], ":") {
				files = strings.Split(userInfo[3], ":")
			} else {
				if len(userInfo) >= 4 {
					files = []string{userInfo[3]}
				}
			}
			if len(userInfo) >=2 {
				user := User{
					Name:        userInfo[0],
					Pword:       userInfo[1],
					Directories: directories,
					Files:       files,
				}
				log.Printf("Allowing User: %v", user)
				users = append(users, user)
				userCount += 1
			}
		}
	}
	log.Printf("Allowing a total of %d users.", userCount)
}

func init() {
	FuncMap := BuildFuncMap()
	fmt.Println("Docroot:", DOCROOT)
	templates = template.Must(template.New("handout").Funcs(FuncMap).ParseGlob(fmt.Sprintf("%s/templates/*", DOCROOT)))
	buildUsers()
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
