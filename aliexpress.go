package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	region := make(chan string)
	go getCountryCodeAli(region)

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}

	headless := false
	devtools := false
	userAgent := "Mozilla/5.0 (Linux; Android 11; IN2020) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.87 Mobile Safari/537.36"

	pw.Chromium.Launch()
	browser, _ := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: &headless,
		Devtools: &devtools,
	})

	context, _ := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: &userAgent,
	})

	page2, _ := context.NewPage()
	page, _ := context.NewPage()

	go page2.Goto(<-region)

	for {
		context.ClearCookies()
		context.ClearPermissions()
		page.Goto("https://twitter.com/i/flow/signup")
		fmt.Println("====================================")
		email := make(chan string)
		code := make(chan string)
		go getEmail(email)

		changeEmail(page)
		notNow(page)
		nameType(page)
		generatedEmail := <-email
		go getTwitterCode(generatedEmail, code)
		emailType(page, generatedEmail)
		dateType(page)
		next(page)
		next(page)
		next(page)

		if !verify(page, <-code) {
			fmt.Println("Code not Found")
			continue
		}

		next(page)

		if isAskedPhone(page) {
			fmt.Println("Phone Number Asked")
			continue
		}

		next(page)

		if !success(page) {
			fmt.Println("Phone Number Asked")
			continue
		}

		fmt.Println("Account Created Successfully")

		page.Goto("https://thirdparty.aliexpress.com/login.htm?spm=a2g0o.home.0.0.650c2145q0XT2s&type=tt&countryCode=US&return_url=https%3A%2F%2Fwww.aliexpress.com%2F")

		if !authCheck(page) {
			fmt.Println("Phone Number Asked")
			continue
		}

		if !loginComplete(page) {
			fmt.Println("Phone Number Asked")
			continue
		}

		page.Goto("https://accounts.aliexpress.com/user/company/change_password_security_prompt.htm")

		if !changePass(page) {
			fmt.Println("Error at ChangePass")
			continue
		}

		aliCode := make(chan string)
		go getAliExpressCode(generatedEmail, aliCode)

		if !sendAliCode(page) {
			fmt.Println("Error at sendAliCode")
			continue
		}

		if !typeAliCode(page, <-aliCode) {
			fmt.Println("Error at typeAliCode")
			continue
		}

		if !typeNewPassword(page) {
			fmt.Println("Error At TypeNewPassword")
			continue
		}

		fmt.Println("Account Created")
		fmt.Println(generatedEmail)
		writeToFile(generatedEmail)
		continue
	}
}

func writeToFile(email string) {
	f, err := os.OpenFile("account.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	emailString := fmt.Sprintf("%s\n", email)
	if _, err := f.WriteString(emailString); err != nil {
		log.Println(err)
	}
}

func typeNewPassword(page playwright.Page) bool {

	if passwordField, err := page.WaitForSelector("input[id='newPwd']"); err != nil {
		errorHandler(err)
	} else {
		passwordField.Type("zoner123")
	}

	if confirmField, err := page.WaitForSelector("input[id='newPwdConfirm']"); err != nil {
		errorHandler(err)
	} else {
		confirmField.Type("zoner123")
	}

	if button, _ := page.WaitForSelector("input[type='submit']"); button != nil {
		button.Click()
		if success, _ := page.WaitForSelector("xpath=//div[contains(@class, 'ui-tipbox-success')]"); success != nil {
			return true
		}
	}

	return false
}

func typeAliCode(page playwright.Page, code string) bool {
	if code == "f" {
		return false
	}

	if passwordField, _ := page.WaitForSelector("[id='J_Checkcode']"); passwordField != nil {
		passwordField.Type(code)
		if button, _ := page.WaitForSelector("[id='submitBtn']"); button != nil {
			button.Click()
			return true
		}
	}
	return false
}

func sendAliCode(page playwright.Page) bool {
	if button, _ := page.WaitForSelector("[id='J_GetCode_Email']"); button != nil {
		button.Click()
		return true
	}
	return false
}

func changePass(page playwright.Page) bool {
	frameElement, _ := page.WaitForSelector("#iframe1")
	frame, _ := frameElement.ContentFrame()
	if button, _ := frame.WaitForSelector("xpath=//*[@id='content']/div/ol/li[1]/a"); button != nil {
		url, _ := button.GetAttribute("href")
		page.Goto(url)
		return true
	}

	return false
}

func loginComplete(page playwright.Page) bool {
	delay := time.Now().Add(time.Second * 45)

	for delay.After(time.Now()) {
		if strings.Contains(page.URL(), "aliexpress.com/?tracelog") {
			fmt.Println("Authorization Successful")
			return true
		}
	}

	return false
}

func authCheck(page playwright.Page) bool {
	quit := make(chan bool)

	go func() {
		if button, _ := page.WaitForSelector("#allow"); button != nil {
			button.Click()
			quit <- true
		}
	}()

	go func() {
		if phone, _ := page.WaitForSelector("[value='Sign In']"); phone != nil {
			quit <- false
		}
	}()

	return <-quit
}

func success(page playwright.Page) bool {
	quit := make(chan bool)

	go func() {
		if uploadPic, _ := page.WaitForSelector("xpath=//*[@id='layers']/div[2]/div/div/div/div/div/div[2]/div[2]/div/div/div[2]/div[2]/div[1]/div/div[2]/span"); uploadPic != nil {
			quit <- true
		}
	}()

	go func() {
		if verify, _ := page.WaitForSelector("xpath=//select"); verify != nil {
			quit <- false
		}
	}()

	return <-quit
}

func isAskedPhone(page playwright.Page) bool {
	quit := make(chan bool)

	go func() {
		if verify, _ := page.WaitForSelector("xpath=//select"); verify != nil {
			quit <- true
		}
	}()

	go func() {
		if password, _ := page.WaitForSelector("xpath=//*[@id='layers']/div[2]/div/div/div/div/div/div[2]/div[2]/div/div/div[2]/div[2]/div[1]/div/div[2]/div/label/div/div[2]/div[1]/input"); password != nil {
			password.Type("zoner123")
			quit <- false
		}
	}()

	return <-quit
}

func verify(page playwright.Page, code string) bool {

	if code == "f" {
		return false
	}

	if verify, err := page.WaitForSelector("xpath=//*[@id='layers']/div[2]/div/div/div/div/div/div[2]/div[2]/div/div/div[2]/div[2]/div[1]/div/div/div[2]/label/div/div[2]/div/input"); err != nil {
		errorHandler(err)
	} else {
		if err := verify.Type(code); err != nil {
			errorHandler(err)
		}
	}
	return true
}

func nameType(page playwright.Page) {
	if name, err := page.WaitForSelector("xpath=//*[@id='layers']/div[2]/div/div/div/div/div/div[2]/div[2]/div/div/div[2]/div[2]/div[1]/div/div[2]/label/div/div[2]/div/input"); err != nil {
		errorHandler(err)
	} else {
		if err := name.Type(getName()); err != nil {
			errorHandler(err)
		}
	}
}

func emailType(page playwright.Page, emailText string) {
	if email, err := page.WaitForSelector("xpath=//*[@id='layers']/div[2]/div/div/div/div/div/div[2]/div[2]/div/div/div[2]/div[2]/div[1]/div/div[3]/label/div/div[2]/div/input"); err != nil {
		errorHandler(err)
	} else {
		if err := email.Type(emailText); err != nil {
			errorHandler(err)
		}
	}
}

func dateType(page playwright.Page) {
	if date, err := page.WaitForSelector("xpath=//*[@id='layers']/div[2]/div/div/div/div/div/div[2]/div[2]/div/div/div[2]/div[2]/div[1]/div/div[5]/div[3]/div/label/div/div[2]/div/input"); err != nil {
		errorHandler(err)
	} else {
		if err := date.Type("03/03/1998"); err != nil {
			errorHandler(err)
		}
	}
}

func notNow(page playwright.Page) {
	if button, _ := page.QuerySelector("//*[@id='layers']/div[3]/div/div/div/div[2]/div[1]"); button != nil {
		clickButton(button)
	}
}

func changeEmail(page playwright.Page) {
	if changeLog, err := page.WaitForSelector("xpath=//*[@id='layers']/div[2]/div/div/div/div/div/div[2]/div[2]/div/div/div[2]/div[2]/div[1]/div/div[4]"); err != nil {
		errorHandler(err)
	} else {
		clickButton(changeLog)
	}
}

func next(page playwright.Page) {
	xpath := "xpath=(//*[@id='layers']/div[2]/div/div/div/div/div/div[2]/div[2]/div/div/div[2]/div[2]/div[2]/div/div[@tabindex='0'] | //*[@id='layers']/div[2]/div/div/div/div/div/div[2]/div[2]/div/div/div[2]/div[2]/div[2]/div[@tabindex='0'] | //*[@id='layers']/div[2]/div/div/div/div/div/div[2]/div[2]/div/div/div[2]/div[2]/div[2]/div[@tabindex='0'] | //*[@id='layers']/div[2]/div/div/div/div/div/div[2]/div[2]/div/div/div[2]/div[2]/div/div/div/div[5][@tabindex='0'])"
	if button, err := page.WaitForSelector(xpath); err != nil {
		errorHandler(err)
	} else {
		clickButton(button)
	}
}

func clickButton(button playwright.ElementHandle) {
	if err := button.Click(); err != nil {
		errorHandler(err)
	}
}

func errorHandler(err error) {
	fmt.Println("Error", err)
}

func getCountryCodeAli(region chan string) {
	data, err := ioutil.ReadFile("country.txt")
	if err != nil {
		log.Panicf("failed reading data from file: %s", err)
	}
	fmt.Printf("Country: %s\n", string(data))
	region <- fmt.Sprintf("https://login.aliexpress.com/setCommonCookie.htm?fromApp=false&currency=EUR&region=%s&bLocale=en_US&site=glo", string(data))
}
