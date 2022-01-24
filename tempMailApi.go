package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	cloudflarebp "github.com/DaRealFreak/cloudflare-bp-go"
	"github.com/sethvargo/go-password/password"
)

func makeRequest(uri string) []byte {
	client := &http.Client{}
	client.Transport = cloudflarebp.AddCloudFlareByPass(client.Transport)

	url := fmt.Sprintf("https://mob1.temp-mail.org/request/%s/format/json", uri)
	res, err := client.Get(url)

	if err != nil {
		errorHandler(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		errorHandler(err)
	}

	return body
}

func domain() string {
	body := makeRequest("domains")
	var domain []string
	if jsonErr := json.Unmarshal(body, &domain); jsonErr != nil {
		errorHandler(jsonErr)
	}

	randomIndex := rand.Intn(len(domain))
	pick := domain[randomIndex]

	return pick
}

func getEmail(email chan string) {
	domain := domain()
	token := strings.ToLower(password.MustGenerate(10, 3, 0, false, false))
	generated := token + domain

	fmt.Println(generated)
	email <- generated
}

func gen_mail_hash(email string) string {
	xor_int := 1573252
	magic_number_java_unicode := 1572864
	var output_list string
	hash := md5.Sum([]byte(email))
	md5Hash := hex.EncodeToString(hash[:])

	for _, letter := range md5Hash {
		output_list += string((int(letter) ^ xor_int) - magic_number_java_unicode)
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(output_list))

	return insert(encoded)
}

func insert(encoded string) string {
	insertData := "%0A"
	first76 := encoded[0:76]
	last := encoded[76:]
	return fmt.Sprintf("%s%s%s%s", first76, insertData, last, insertData)
}

func get_message(email string) []byte {
	uri := fmt.Sprintf("mail/id/%s", gen_mail_hash(email))
	return makeRequest(uri)
}

func getTwitterCode(email string, code chan string) {
	delay := time.Now().Add(time.Second * 30)
	for delay.After(time.Now()) {
		body := get_message(email)
		response := string(body)
		re2, _ := regexp.Compile("\"[0-9][0-9][0-9][0-9][0-9][0-9][[:space:]]")
		result := re2.FindAllString(response, 1)
		if len(result) > 0 {
			code <- result[0][1 : len(result[0])-1]
		}
	}

	code <- "f"
}

func getAliExpressCode(email string, code chan string) {
	delay := time.Now().Add(time.Second * 30)
	for delay.After(time.Now()) {
		body := get_message(email)
		response := string(body)
		re2, _ := regexp.Compile(">[0-9][0-9][0-9][0-9][0-9][0-9]<")
		result := re2.FindAllString(response, 1)

		if len(result) > 0 {
			code <- result[0][1 : len(result[0])-1]
		}
	}

	code <- "f"
}

func getName() string {
	return password.MustGenerate(10, 3, 0, true, false)
}
