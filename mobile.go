package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func buildPayload(msg string, mobile string) map[string]string {
	userid := os.Getenv("SMS_OTP_USERID")
	password := os.Getenv("SMS_OTP_PASSWORD")

	payload := map[string]string{
		"method":       "sendMessage",
		"msg_type":     "TEXT",
		"auth_scheme":  "PLAIN",
		"v":            "1.1",
		"format":       "text",
		"override_dnd": "true",
		"userid":       userid,
		"password":     password,
		"msg":          msg,
		"send_to":      mobile,
	}

	return payload
}

func SendOtpSms(mobile string, otpCode string) {
	msg := fmt.Sprintf("%s is your OTP Verification Code. Software by MITSHI India Ltd - Listed on BSE since 28 years - %s", otpCode, "https://mitshi.in")
	payload := buildPayload(msg, mobile)

	params := url.Values{}
	for k, v := range payload {
		params.Add(k, v)
	}

	gupshupApi := "http://enterprise.smsgupshup.com/GatewayAPI/rest"
	resp, err := http.PostForm(gupshupApi, params)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(err)
		return
	}

	log.Print("Response for gupshup", string(body))
}
