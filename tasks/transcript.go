package tasks

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"
)

type TranscriptTask struct {
	task              *Task
	RelayState        string
	SAMLResponse      string
	SAMLRequest       string
	Name              string
	UserId            string
	Degree            string
	DegreeDescription string
	SchoolKey         string
	SchoolDescription string
	AuditInfo         []AuditInfo
}

func (t *TranscriptTask) VisitHomepage() error {
	fmt.Println("Visiting homepage")

	request, err := http.NewRequest(http.MethodGet, "https://dw-prod.ec.fhda.edu/responsiveDashboard/worksheets/WEB31", nil)
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("user-agent", t.task.UserAgent)

	resp, err := t.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}

	if len(body) > 0 {
	}
	return nil
}

func (t *TranscriptTask) Login() error {
	fmt.Println("Logging in")
	loginData := fmt.Sprintf("j_username=%s&j_password=%s&_eventId_proceed=", t.task.Username, t.task.Password)
	request, err := http.NewRequest(http.MethodPost, "https://ssoshib.fhda.edu/idp/profile/SAML2/Redirect/SSO?execution=e1s1", bytes.NewBufferString(loginData))
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	request.Header.Add("user-agent", t.task.UserAgent)

	resp, err := t.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}
	reader := strings.NewReader(string(body))
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return UnableToGetDocument
	}
	var message string
	document.Find("div[class='alert alert-danger']").Each(func(index int, element *goquery.Selection) {
		message = strings.TrimSpace(element.Text())
	})

	switch message {
	case "The password you entered was incorrect.":
		return InvalidCredentials
	case "You may be seeing this page because you used the Back button while browsing a secure web site or application. Alternatively, you may have mistakenly bookmarked the web login form instead of the actual web site you wanted to bookmark or used a link created by somebody else who made the same mistake.  Left unchecked, this can cause errors on some browsers or result in you returning to the web site you tried to leave, so this page is presented instead.":
		return SessionCorrupted
	case "":
		break
	default:
		return errors.New(message)
	}

	relayStateValue := ""
	samlResponseValue := ""
	document.Find("input[name='SAMLResponse']").Each(func(index int, element *goquery.Selection) {
		value, exists := element.Attr("value")
		if exists {
			samlResponseValue = value
		}
	})
	document.Find("input[name='RelayState']").Each(func(index int, element *goquery.Selection) {
		value, exists := element.Attr("value")
		if exists {
			relayStateValue = value
		}
	})

	if len(samlResponseValue) == 0 {
		return NoSamlResponseValue
	}

	t.RelayState = relayStateValue
	t.SAMLResponse = samlResponseValue

	return nil
}

func (t *TranscriptTask) SubmitCommonAuth() error {
	fmt.Println("Submitting Common Auth SSO")

	values := url.Values{
		"RelayState":   {t.RelayState},
		"SAMLResponse": {t.SAMLResponse},
	}

	request, err := http.NewRequest(http.MethodPost, "https://eis-prod.ec.fhda.edu/commonauth", bytes.NewBufferString(values.Encode()))
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	request.Header.Add("user-agent", t.task.UserAgent)

	resp, err := t.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}
	reader := strings.NewReader(string(body))
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return UnableToGetDocument
	}

	samlResponseValue := ""
	document.Find("input[name='SAMLResponse']").Each(func(index int, element *goquery.Selection) {
		value, exists := element.Attr("value")
		if exists {
			samlResponseValue = value
		}
	})

	if len(samlResponseValue) == 0 {
		return NoSamlResponseValue
	}

	t.SAMLResponse = samlResponseValue
	return nil
}

func (t *TranscriptTask) SubmitSSO() error {
	fmt.Println("Submitting SSO")

	values := url.Values{
		"RelayState":   {t.RelayState},
		"SAMLResponse": {t.SAMLResponse},
	}

	request, err := http.NewRequest(http.MethodPost, "https://dw-prod.ec.fhda.edu/responsiveDashboard/saml/SSO", bytes.NewBufferString(values.Encode()))
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	request.Header.Add("user-agent", t.task.UserAgent)

	resp, err := t.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}
	if len(body) > 0 {
	}
	return nil
}

func (t *TranscriptTask) GetUserInfo() error {
	fmt.Println("Getting user info")

	request, err := http.NewRequest(http.MethodGet, "https://dw-prod.ec.fhda.edu/responsiveDashboard/api/students/myself", nil)
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("user-agent", t.task.UserAgent)

	resp, err := t.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}

	if len(body) > 0 {
		userInfo := UserInfo{}
		if err := json.Unmarshal(body, &userInfo); err != nil {
			fmt.Println(err)
			return UnableToParseJSON
		}
		for _, student := range userInfo.Embedded.Students {
			t.Name = student.Name
			t.UserId = student.ID
			t.SchoolKey = student.Goals[0].School.Key
			t.SchoolDescription = student.Goals[0].School.Description
			t.Degree = student.Goals[0].Degree.Key
			t.DegreeDescription = student.Goals[0].Degree.Description
		}
	}
	return nil
}

func (t *TranscriptTask) GetAudit() error {
	fmt.Printf("Getting audit for %s - %s\n", t.Name, t.SchoolDescription)

	url := fmt.Sprintf("https://dw-prod.ec.fhda.edu/responsiveDashboard/api/audit?studentId=%s&school=%s&degree=%s&is-process-new=false&audit-type=AA&auditId=&include-inprogress=true&include-preregistered=true&aid-term=",
		t.UserId,
		t.SchoolKey,
		t.Degree)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("user-agent", t.task.UserAgent)

	resp, err := t.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}

	if len(body) > 0 {
		var auditInfo []AuditInfo
		audit := Audit{}
		if err := json.Unmarshal(body, &audit); err != nil {
			fmt.Println(err)
			return UnableToParseJSON
		}
		for _, class := range audit.ClassInformation.ClassArray {
			classInfo := AuditInfo{
				Term:        class.TermLiteralLong,
				Section:     class.Discipline,
				Number:      class.Number,
				CourseTitle: class.CourseTitle,
				LetterGrade: class.LetterGrade,
				Credits:     class.Credits,
			}
			auditInfo = append(auditInfo, classInfo)
		}
		t.AuditInfo = auditInfo
	}
	return nil
}

func (t *TranscriptTask) ExportTranscript() error {
	fmt.Println("Exporting transcript")

	currentTime := time.Now()
	fileName := fmt.Sprintf("%s-%s-%s.csv", t.Name, t.Degree, currentTime.Format("2006-01-02_15-04-05"))
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Term", "Section", "Number", "Course Title", "Letter Grade", "Credits"}
	fmt.Printf("Writing to %s\n", fileName)
	err = writer.Write(header)
	if err != nil {
		return FailedToWrite
	}
	for _, audit := range t.AuditInfo {
		record := []string{
			audit.Term,
			audit.Section,
			audit.Number,
			audit.CourseTitle,
			audit.LetterGrade,
			audit.Credits,
		}
		err = writer.Write(record)
		if err != nil {
			return FailedToWrite
		}
	}
	fmt.Println("Exported transcript data")
	return nil
}

func (t *TranscriptTask) Run() error {
	steps := []func() error{
		t.VisitHomepage,
		t.Login,
		t.SubmitCommonAuth,
		t.SubmitSSO,
		t.GetUserInfo,
		t.GetAudit,
		t.ExportTranscript,
	}

	for _, step := range steps {
		if err := Retry(t.task.RetryAmount, t.task.RetryDuration, step); err != nil {
			return MaximumAttemptsRetry
		}
	}

	t.task.Client.CloseIdleConnections()
	return nil
}

func NewTranscriptTask(task *Task) *TranscriptTask {
	return &TranscriptTask{task: task}
}
