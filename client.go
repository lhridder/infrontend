package infrontend

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

func GetClientHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./views/client/index.html", "./views/client/header.html", "./views/footer.html")
	if err != nil {
		log.Printf("Failed to parse template: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	user, err := GetUserFromRequest(r)
	if err != nil {
		log.Printf("Failed to get user from request: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	data := pageData{
		Year:    string(rune(time.Now().Year())),
		Version: Version,
		Title:   "Client",
		Script:  "client.js",
		User:    user,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func GetLogout(w http.ResponseWriter, r *http.Request) {

}
