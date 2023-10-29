package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/joho/godotenv"
	"github.com/veil/tasks"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading env file: %s\n", err)
		return
	}

	username := os.Getenv("CAMPUSID")
	password := os.Getenv("PASSWORD")
	mode := os.Getenv("MODE")
	subject := os.Getenv("SUBJECT")
	year := os.Getenv("YEAR")
	quarter := strings.ToLower(os.Getenv("QUARTER"))
	campus := strings.ToLower(os.Getenv("CAMPUS"))
	crntoadd := os.Getenv("CRNTOADD")
	retryamount := os.Getenv("RETRY_AMOUNT")
	retryduration := os.Getenv("RETRY_DURATION")
	webhookURL := os.Getenv("DISCORD_WEBHOOK")

	t := &tasks.Task{}

	jar := tls_client.NewCookieJar()
	client_options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_117),
		tls_client.WithCookieJar(jar),
	}
	t.Client, _ = tls_client.NewHttpClient(tls_client.NewLogger(), client_options...)
	t.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"
	t.Username = username
	t.Password = password
	t.Subject = subject

	retryAmount, err := strconv.Atoi(retryamount)
	retryDuration, err := strconv.Atoi(retryduration)

	t.RetryAmount = retryAmount
	t.RetryDuration = time.Duration(retryDuration * int(time.Second))
	t.CourseNumber = crntoadd
	t.WebhookURL = webhookURL

	if mode == "SEARCH" || mode == "SIGNUP" {
		yearint, err := strconv.Atoi(year)
		if err != nil {
			fmt.Println(err)
			return
		}
		termId, err := tasks.BuildTermId(yearint, campus, quarter)
		t.TermId = termId

		termDesc, err := t.SearchTerm()
		if err != nil {
			fmt.Printf("Warning: Term not found\n")
		} else {
			fmt.Printf("Found term: %s\n", termDesc)
		}

		if mode == "SEARCH" {
			search := tasks.NewSearchTask(t)
			if err := search.Run(); err != nil {
				fmt.Println(err)
			}
		} else if mode == "SIGNUP" {
			signup := tasks.NewSignupTask(t)
			if err := signup.Run(); err != nil {
				fmt.Println(err)
			}
		}
	} else if mode == "TRANSCRIPT" {
		transcript := tasks.NewTranscriptTask(t)
		if err := transcript.Run(); err != nil {
			fmt.Println(err)
		}
	}
	for {
		time.Sleep(time.Second)
	}
}
