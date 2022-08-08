package infrontend

import (
	"github.com/google/uuid"
	"net/http"
)

type User struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Hash      string `json:"hash"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Suspended bool   `json:"suspended"`
}

type APItoken struct {
	User        uuid.UUID `json:"user"`
	WriteAccess bool      `json:"writeAccess"`
}

type Session struct {
	User uuid.UUID `json:"user"`
}

func GetClientHome(w http.ResponseWriter, r *http.Request) {

}
