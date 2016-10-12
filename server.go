package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	//"time"
	//"os"

	"github.com/gorilla/mux"
)

//========ROUTES========
func main() {

	r := mux.NewRouter()
	r.HandleFunc("/list", Auth(ListFilesHandler))
	r.HandleFunc("/edit", Auth(EditFileHandler))

	err := http.ListenAndServe("localhost:7777", r)
	if err != nil {
		log.Fatal(err)
	}
}

//----------------------

//=======HANDLERS=======
func ListFilesHandler(w http.ResponseWriter, r *http.Request, u User) {
	for _, d := range u.Directories {
		files, err := ioutil.ReadDir(d)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			w.Write([]byte(fmt.Sprintf("<a href=\"/edit?filepath=%[1]s/%[2]s\">File -> %[2]s<\a><br>", d, f.Name())))
		}
	}
}

func EditFileHandler(w http.ResponseWriter, r *http.Request, u User) {
	vals := r.URL.Query()
	filepath := vals["filepath"][0]

	if u.CanEditFile(filepath) {
		bytes, err := ioutil.ReadFile(filepath)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(bytes)
		return
	}
	w.Write([]byte("You are not allowed to edit this file."))
	//TODO: User templates or mustache or some tool to write html to the client instead of these bytes.
}

//-------------------------------

func GetUser(r *http.Request) User {
	u := User{
		Name:        "Gabe",
		Pword:       "Hughes",
		Directories: []string{"userfiles"},
		Files:       []string{"userfiles/goober.txt"},
	}
	return u
}

//=======AUTHENTICATION=======

type AuthedHandlerFunc func(w http.ResponseWriter, r *http.Request, u User)

//Auth
//This ensures that only authed users can access a handler (endpoint).
//It takes in a custom handler with an extra parameter (user) and fills that information in.
//It then converts it to a function that golang can associate with an endpoint. (HandlerFunc)
//Wrapping the handlers this way takes the authentication logic out of each individual endpoint.
func Auth(h AuthedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := GetUser(r)
		if ValidUser(u) {
			h(w, r, u)
		}
		return
	}
}

//--------------------------
