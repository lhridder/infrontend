package infrontend

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type User struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Hash      string `json:"hash"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Suspended bool   `json:"suspended"`
	Admin     bool   `json:"admin"`
}

type APItoken struct {
	User        uuid.UUID `json:"user"`
	WriteAccess bool      `json:"writeAccess"`
}

type Session struct {
	User uuid.UUID `json:"user"`
}

func GetRegister(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./views/register.html", "./views/header.html", "./views/footer.html")
	if err != nil {
		log.Printf("Failed to parse template: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	data := pageData{
		Year:    string(rune(time.Now().Year())),
		Version: "0.1-beta",
		Title:   "Register",
		Script:  "register.js",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func GetLogin(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./views/login.html", "./views/header.html", "./views/footer.html")
	if err != nil {
		log.Printf("Failed to parse template: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	data := pageData{
		Year:    string(rune(time.Now().Year())),
		Version: Version,
		Title:   "Login",
		Script:  "login.js",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func PostRegister(w http.ResponseWriter, r *http.Request) {
	rawData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	type register struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var registration register
	err = json.Unmarshal(rawData, &registration)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	if !govalidator.IsEmail(registration.Email) || !govalidator.IsAlphanumeric(registration.Username) {
		log.Println("invalid mail or username")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	target, _, err := FindUser(registration.Username)
	if !reflect.ValueOf(target).IsZero() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("username taken"))
		return
	}

	target, _, err = FindUser(registration.Email)
	if !reflect.ValueOf(target).IsZero() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("email taken"))
		return
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(registration.Password), 14)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	var user User
	user.Email = registration.Email
	user.Username = registration.Username
	user.Hash = string(bytes)

	//TODO implement email check and suspend

	err = StoreUser(user)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func PostLogin(w http.ResponseWriter, r *http.Request) {
	rawData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	type login struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var Login login
	err = json.Unmarshal(rawData, &Login)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	user, useruuid, err := FindUser(Login.Username)

	if reflect.ValueOf(user).IsZero() {
		log.Println("no user found")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(Login.Password))
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	session := Session{
		User: useruuid,
	}

	sessionuuid, err := StoreSession(session)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionuuid.String(),
		Expires: time.Now().Add(expiration),
	})
}

func Auth() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			token, err := bearerToken(r)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			if token != "" {
				apitoken, err := GetToken(token)
				if err != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				if reflect.ValueOf(apitoken).IsZero() {
					http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}

				if !apitoken.WriteAccess {
					if r.Method != "" {
						http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
						return
					}
				}

				user, err := GetUser(apitoken.User)
				if err != nil {
					return
				}

				ctx := context.WithValue(context.Background(), "user", user)
				r.WithContext(ctx)

				next.ServeHTTP(w, r)
			}

			cookie, err := r.Cookie("session_token")
			if err == http.ErrNoCookie {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			sessionuuid, err := uuid.Parse(cookie.Value)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			session, err := GetSession(sessionuuid)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if reflect.ValueOf(session).IsZero() {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			requestcontext := context.WithValue(r.Context(), "user", session.User)
			next.ServeHTTP(w, r.WithContext(requestcontext))
		}
		return http.HandlerFunc(fn)
	}
}

func Admin() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserFromRequest(r)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if !user.Admin {
				http.Redirect(w, r, "/client", http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func GetUserFromRequest(r *http.Request) (User, error) {
	var useruuid uuid.UUID
	var ok bool
	if useruuid, ok = r.Context().Value("user").(uuid.UUID); !ok {
		return User{}, err
	}

	return GetUser(useruuid)
}

func bearerToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", nil
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 {
		return "", errors.New("token with incorrect bearer format")
	}

	token := strings.TrimSpace(parts[1])
	return token, nil
}
