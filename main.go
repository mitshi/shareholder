package main

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mitshi/shareholder/internal/db"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

var DB *sql.DB
var Q *db.Queries
var C *OtpCache

var t = template.Must(template.ParseGlob("./templates/" + "*.html"))

var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type homeForm struct {
	Name              string
	Email             string
	Mobile            string
	FolioNumber       string
	CertificateNumber string
	PanNumber         string
	AgreeTerms        bool
}

func (form *homeForm) validate() (bool, map[string]string) {
	formErrors := make(map[string]string)
	failed := false

	form.Name = strings.TrimSpace(form.Name)
	form.Email = strings.TrimSpace(form.Email)
	form.Mobile = strings.TrimSpace(form.Mobile)

	if form.Name == "" || form.Email == "" || form.Mobile == "" {
		formErrors["Empty Field"] = "Please fill all form fields."
		failed = true
	}

	if len(form.Mobile) != 10 {
		formErrors["Mobile Number"] = "Please use a 10 Digit mobile number."
		failed = true
	}

	match := rxEmail.Match([]byte(form.Email))
	if match == false {
		formErrors["Email"] = "Please enter a valid email address."
		failed = true
	}

	return failed, formErrors
}

func verifyCompleteHandler(w http.ResponseWriter, r *http.Request) {
	var ctx interface{}
	t.ExecuteTemplate(w, "verify_complete.html", ctx)
}

func verifyMobileHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	shareholderId := session.Values["id"].(int32)
	shareholder, err := Q.GetShareholder(context.TODO(), shareholderId)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}
	otpCode, _ := C.GetOtp(shareholder.Mobile)

	if r.Method == "GET" {
		var ctx interface{}
		t.ExecuteTemplate(w, "verify_mobile.html", ctx)
		return
	} else {
		enteredOtp := r.FormValue("otpCode")
		if enteredOtp != otpCode {
			formErrors := map[string]string{
				"OTP": "Entered OTP is not correct or has expired. Please try again.",
			}
			t.ExecuteTemplate(w, "verify_mobile.html", formErrors)
			return
		}

		params := db.UpdateMobileVerifyShareholderParams{
			ID:             shareholderId,
			MobileVerified: true,
		}
		Q.UpdateMobileVerifyShareholder(context.TODO(), params)

		http.Redirect(w, r, "/verify/complete/", http.StatusSeeOther)
		return
	}
}

func verifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	shareholderId := session.Values["id"].(int32)
	shareholder, err := Q.GetShareholder(context.TODO(), shareholderId)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}
	otpCode, _ := C.GetOtp(shareholder.Email)

	log.Println("Shareholder ID: " + strconv.FormatInt(int64(shareholderId), 10))

	if r.Method == "GET" {
		log.Println("Email handler OTP for email " + shareholder.Email)
		log.Println(otpCode)
		var ctx interface{}
		t.ExecuteTemplate(w, "verify_email.html", ctx)
		return
	} else {
		enteredOtp := r.FormValue("otpCode")
		if enteredOtp != otpCode {
			formErrors := map[string]string{
				"OTP": "Entered OTP is not correct or has expired. Please try again.",
			}
			t.ExecuteTemplate(w, "verify_email.html", formErrors)
			return
		}

		params := db.UpdateEmailVerifyShareholderParams{
			ID:            shareholderId,
			EmailVerified: true,
		}
		shareholder, _ := Q.UpdateEmailVerifyShareholder(context.TODO(), params)
		mobileOtpCode := C.CreateOtp(shareholder.Mobile)
		go SendOtpSms(shareholder.Mobile, mobileOtpCode)

		http.Redirect(w, r, "/verify/mobile/", http.StatusSeeOther)
		return
	}
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var ctx interface{}
		t.ExecuteTemplate(w, "home.html", ctx)
	} else {
		form := homeForm{
			Name:              r.FormValue("name"),
			Email:             r.FormValue("email"),
			Mobile:            r.FormValue("mobile"),
			FolioNumber:       r.FormValue("folioNumber"),
			CertificateNumber: r.FormValue("certificateNumber"),
			PanNumber:         r.FormValue("panNumber"),
			AgreeTerms:        true,
		}
		failed, formErrors := form.validate()
		if failed {
			t.ExecuteTemplate(w, "home.html", formErrors)
			return
		}

		params := db.CreateShareholderParams(form)
		shareholder, _ := Q.CreateShareholder(context.TODO(), params)
		otpCode := C.CreateOtp(form.Email)
		log.Println("OTP Created is: " + otpCode)
		go SendMyMail(form.Email, otpCode)

		log.Println(shareholder)

		session, _ := store.Get(r, "session-name")
		session.Values["id"] = shareholder.ID
		session.Save(r, w)

		http.Redirect(w, r, "/verify/email/", http.StatusSeeOther)
		return
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var err error
	dbConnection, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln(err)
	}
	if err = dbConnection.Ping(); err != nil {
		log.Fatalln(err)
	}

	// global
	DB = dbConnection
	Q = db.New(dbConnection)
	C = NewOtpCache()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// TODO: use csrf middleware

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.HandleFunc("/", homePageHandler)
	r.HandleFunc("/verify/email/", verifyEmailHandler)
	r.HandleFunc("/verify/mobile/", verifyMobileHandler)
	r.HandleFunc("/verify/complete/", verifyCompleteHandler)
	r.Handle("/static/", http.FileServer(http.Dir("static")))

	serverUrl := os.Getenv("SERVER_URL")
	log.Println("Starting server on " + serverUrl)
	log.Fatal(http.ListenAndServe(serverUrl, r))
}
