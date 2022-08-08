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

type pageData struct {
	Year    string
	Version string
	Title   string
	Script  string
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
		Version: "0.1-beta",
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

	target, err := findUser(registration.Username)
	if !reflect.ValueOf(target).IsZero() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("username taken"))
		return
	}

	target, err = findUser(registration.Email)
	if !reflect.ValueOf(target).IsZero() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("email taken"))
		return
	}

	log.Println(registration.Password)
	bytes, err := bcrypt.GenerateFromPassword([]byte(registration.Password), 14)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	var user User
	user.Email = registration.Email
	user.Username = registration.Username
	user.Hash = string(bytes)
	log.Println(user.Hash)

	serializedUser, err := json.Marshal(user)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	useruuid := uuid.New()
	_, err = rdb.Set(ctx, "user:"+useruuid.String(), string(serializedUser), 0).Result()
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

	user, err := findUser(Login.Username)

	if reflect.ValueOf(user).IsZero() {
		log.Println("no user found")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(Login.Password))
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	//TODO give cookie

}

func findUser(search string) (User, error) {
	var user User

	users, err := rdb.Keys(ctx, "user:*").Result()
	if err != nil {
		return user, err
	}

	for _, item := range users {
		useruuid, err := uuid.Parse(strings.Split(item, "user:")[1])
		user, err = GetUser(useruuid)
		if err != nil {
			return User{}, err
		}
		if user.Email == search || user.Username == search {
			break
		}
	}

	return user, nil
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

			session, err := GetSession(cookie.Value)
			if err != nil {

			}

			if reflect.ValueOf(session).IsZero() {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			user, err := GetUser(session.User)
			if err != nil {
				return
			}

			ctx := context.WithValue(context.Background(), "user", user)
			r.WithContext(ctx)

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
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
