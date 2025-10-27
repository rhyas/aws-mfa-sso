package auth

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"github.com/rhyas/aws-mfa-sso/internal/ui"
)

// PerformLogin performs the headless SSO login using the provided URL
func PerformLogin(url string) {
	browser := rod.New().
		MustConnect().
		Trace(false)

	loadCookies(*browser)

	defer browser.MustClose()

	err := rod.Try(func() {
		page := browser.MustPage(url)

		// authorize
		page.MustElementR("button", "Next").MustWaitEnabled().MustPress()

		// sign-in
		page.Race().ElementR("button", "Allow").MustHandle(func(e *rod.Element) {
		}).Element("#awsui-input-0").MustHandle(func(e *rod.Element) {
			signIn(*page)
		}).MustDo()

		// After sign-in, confirm the code is correct
		log.Println("Waiting for code confirmation page...")
		page.MustElementR("button", "Confirm and continue").MustWaitEnabled().MustClick()
		log.Println("Code confirmed")

		// Wait for device approval page (without navigation, SPA changes)
		log.Println("Waiting for device approval page...")
		page.MustElementR("button", "Allow").MustWaitEnabled().MustClick()
		log.Println("Device access allowed")

		// Wait for "Request approved" confirmation
		log.Println("Waiting for approval confirmation...")
		page.MustElementR("div", "Request approved").MustWaitLoad()
		log.Println("Request approved!")

		saveCookies(*browser)
	})

	if errors.Is(err, context.DeadlineExceeded) {
		log.Panic("Timeout")
	} else if err != nil {
		log.Panic(err)
	}
}

// signIn executes the AWS SSO signin step
func signIn(page rod.Page) {
	// Get user credentials
	username := ui.PromptUsername()
	password := ui.PromptPassword()

	// Fill in username and password
	page.MustElement("#awsui-input-0").MustInput(username).MustPress(input.Enter)
	page.MustElement("#awsui-input-1").MustInput(password).MustPress(input.Enter)

	// Get MFA code
	mfaCode := ui.PromptMFA()

	// Input the MFA code into the MFA field
	page.MustElement("#awsui-input-2").MustInput(mfaCode).MustPress(input.Enter)
	log.Println("MFA code submitted")
}

// loadCookies loads saved cookies from the home directory
func loadCookies(browser rod.Browser) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}

	data, _ := os.ReadFile(dirname + "/.aws-mfa-sso")
	sEnc, _ := b64.StdEncoding.DecodeString(string(data))
	var cookie *proto.NetworkCookie
	json.Unmarshal(sEnc, &cookie)

	if cookie != nil {
		browser.MustSetCookies(cookie)
	}
}

// saveCookies saves the authentication cookie to the home directory
func saveCookies(browser rod.Browser) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}

	cookies := browser.MustGetCookies()

	for _, cookie := range cookies {
		if cookie.Name == "x-amz-sso_authn" {
			data, _ := json.Marshal(cookie)

			sEnc := b64.StdEncoding.EncodeToString([]byte(data))
			err = os.WriteFile(dirname+"/.aws-mfa-sso", []byte(sEnc), 0644)

			if err != nil {
				log.Panic(err)
			}
			break
		}
	}
}
