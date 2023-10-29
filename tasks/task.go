package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

type Task struct {
	Subject       string
	Term          string
	TermId        string
	CoursesToAdd  []string
	Client        tls_client.HttpClient
	UserAgent     string
	RetryDuration time.Duration
	RetryAmount   int
	Username      string
	Password      string
	WebhookURL    string
}

func BuildTermId(year int, campus string, quarter string) (string, error) {
	campusCode, ok1 := CampusCodes[campus]
	if !ok1 {
		return "", InvalidCampus
	}
	quarterCode, ok2 := QuarterCodes[quarter]
	if !ok2 {
		return "", InvalidQuarter
	}
	return fmt.Sprintf("%d%d%d", year, quarterCode, campusCode), nil
}

func (task *Task) SearchTerm() (string, error) {
	fmt.Println("Searching for term")

	request, err := http.NewRequest(http.MethodGet, "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/classSearch/getTerms?searchTerm=&offset=1&max=10", nil)
	if err != nil {
		return "", FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("user-agent", task.UserAgent)

	resp, err := task.Client.Do(request)
	if err != nil {
		return "", FailedToMakeRequest
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", FailedToReadResponseBody
	}

	terms := Terms{}
	if err := json.Unmarshal(body, &terms); err != nil {
		fmt.Println(err)
		return "", UnableToParseJSON
	}

	var termDesc string
	for _, term := range terms {
		if task.TermId == term.Code {
			termDesc = term.Description
		}
	}
	if len(termDesc) > 0 {
		return termDesc, nil
	} else {
		return "", TermNotFound
	}
}

func Convert24HourTimeTo12HourFormat(input string) string {
	if len(input) != 4 {
		return ""
	}

	hour, err := strconv.Atoi(input[:2])
	if err != nil {
		return ""
	}

	if hour < 0 || hour > 23 {
		return ""
	}

	minutes := input[2:]

	if hour == 0 {
		return fmt.Sprintf("12:%s AM", minutes)
	} else if hour == 12 {
		return fmt.Sprintf("12:%s PM", minutes)
	} else if hour < 12 {
		return fmt.Sprintf("%d:%s AM", hour, minutes)
	} else {
		return fmt.Sprintf("%d:%s PM", hour-12, minutes)
	}
}

func extractModel(jsonData []byte) (map[string]interface{}, error) {
	var data AddCourse
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}
	return data.Model, nil
}

func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			fmt.Println(err)
			time.Sleep(sleep)
		}
		err = f()
		if err == nil {
			return nil
		} else if err == CourseSearchUnsuccessful {
			fmt.Println(err)
			break
		} else if err == NotEligibleToRegister {
			fmt.Println(err)
			break
		} else if err == FailedSubmittingChangesCRNErrors {
			fmt.Println(err)
			break
		}
	}
	return MaximumAttemptsRetry
}

func formatDuration(time time.Duration) string {
	totalSeconds := int64(time.Seconds())

	days := totalSeconds / (60 * 60 * 24)
	hours := (totalSeconds % (60 * 60 * 24)) / (60 * 60)
	minutes := (totalSeconds % (60 * 60)) / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
}

func (task *Task) sendSuccessfulEnrollmentNotification(CourseTitle string) error {

	if len(task.WebhookURL) < 0 {
		return NoWebHookURL
	}
	now := time.Now().UTC()
	payload := WebhookPayload{
		Username: "Veil",
		Embeds: []Embed{
			{
				Title:       "Successful Enrollment",
				Color:       5814783,
				Description: CourseTitle,
				Footer: &Footer{
					Text: "Veil",
				},
				Timestamp: now.Format("2006-01-02T15:04:05.000Z"),
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return UnableToParseJSON
	}

	request, err := http.NewRequest(http.MethodPost, task.WebhookURL, bytes.NewBufferString(string(jsonData)))
	if err != nil {
		return FailedToMakeRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("content-type", "application/json")
	request.Header.Add("user-agent", task.UserAgent)

	resp, err := task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()
	if resp.StatusCode <= 201 {
		return FailedToSendNotification
	}
	return nil
}
