package infrontend

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

func GetAdminHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./views/admin/index.html", "./views/admin/header.html", "./views/footer.html")
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
		Title:   "Admin",
		Script:  "admin.js",
		User:    user,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
