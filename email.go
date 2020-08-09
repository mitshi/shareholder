package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"

	"github.com/jordan-wright/email"
	"github.com/matcornic/hermes/v2"
)

func SendMyMail(toEmail string, otpCode string) {
	fmt.Println("Sending Mail...")
	// Configure hermes by setting a theme and your product info
	h := hermes.Hermes{
		// Optional Theme
		// Theme: new(Default)
		Product: hermes.Product{
			// Appears in header & footer of e-mails
			Name: "MITSHI INDIA LTD",
			Link: "https://mitshi.in",
			// Optional product logo
			Logo: "https://mitshi.in/static/img/mitshi-purple-rect-logo.png",
		},
	}

	hEmail := hermes.Email{
		Body: hermes.Body{
			Name: "Shareholder email update.",
			Intros: []string{
				"Please confirm your email address.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Please click here",
					Button: hermes.Button{
						Color: "#22BC66", // Optional action button color
						Text:  "Confirm",
						Link:  "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010",
					},
				},
			},
			Outros: []string{
				"",
			},
		},
	}

	// Generate an HTML email with the provided contents (for modern clients)
	emailBody, err := h.GenerateHTML(hEmail)
	if err != nil {
		log.Fatalln(err) // Tip: Handle error with something else than a panic ;)
	}

	// Generate the plaintext version of the e-mail (for clients that do not support xHTML)
	emailText, err := h.GeneratePlainText(hEmail)
	if err != nil {
		log.Fatalln(err) // Tip: Handle error with something else than a panic ;)
	}

	e := &email.Email{
		To:      []string{toEmail},
		From:    "no-reply@transactions.mitshi.in",
		Subject: "OTP Code: " + otpCode,
		Text:    []byte(emailText),
		HTML:    []byte(emailBody),
	}
	err = e.Send("smtp.sendgrid.net:587", smtp.PlainAuth("", "apikey", os.Getenv("SMTP_PASS"), "smtp.sendgrid.net"))
	if err != nil {
		log.Fatalln(err)
	}
}
