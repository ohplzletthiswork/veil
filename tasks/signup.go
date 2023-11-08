package tasks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"
)

type SignupTask struct {
	task         *Task
	AuthToken    string
	RelayState   string
	SAMLResponse string
	SAMLRequest  string
	Model        map[string]interface{}
}

func (s *SignupTask) VisitHomepage() error {
	fmt.Println("Visiting homepage")

	request, err := http.NewRequest(http.MethodGet, "https://ssb-prod.ec.fhda.edu/ssomanager/saml/login?relayState=%2Fc%2Fauth%2FSSB%3Fpkg%3Dhttps%3A%2F%2Fssb-prod.ec.fhda.edu%2FPROD%2Ffhda_uportal.P_DeepLink_Post%3Fp_page%3Dbwskfreg.P_AltPin%26p_payload%3De30%3D", nil)
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return UnknownHTTPResponseStatus
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}
	if len(body) > 0 {
	}
	return nil
}

func (s *SignupTask) Login() error {
	fmt.Println("Logging in")

	loginData := fmt.Sprintf("j_username=%s&j_password=%s&_eventId_proceed=", s.task.Username, s.task.Password)
	request, err := http.NewRequest(http.MethodPost, "https://ssoshib.fhda.edu/idp/profile/SAML2/Redirect/SSO?execution=e1s1", bytes.NewBufferString(loginData))
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return UnknownHTTPResponseStatus
	}

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
	s.RelayState = relayStateValue
	s.SAMLResponse = samlResponseValue
	return nil
}

func (s *SignupTask) SubmitCommonAuth() error {
	fmt.Println("Submitting Common Auth SSO")

	values := url.Values{
		"RelayState":   {s.RelayState},
		"SAMLResponse": {s.SAMLResponse},
	}

	request, err := http.NewRequest(http.MethodPost, "https://eis-prod.ec.fhda.edu/commonauth", bytes.NewBufferString(values.Encode()))
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return UnknownHTTPResponseStatus
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}
	reader := strings.NewReader(string(body))
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return UnableToGetDocument
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
	s.RelayState = relayStateValue
	s.SAMLResponse = samlResponseValue
	return nil
}

func (s *SignupTask) SubmitSSOManager() error {
	fmt.Println("Submitting SSO Manager")

	values := url.Values{
		"RelayState":   {s.RelayState},
		"SAMLResponse": {s.SAMLResponse},
	}

	request, err := http.NewRequest(http.MethodPost, "https://ssb-prod.ec.fhda.edu/ssomanager/saml/SSO", bytes.NewBufferString(values.Encode()))
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
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

func (s *SignupTask) RegisterPostSignIn() error {
	fmt.Println("Registering post sign in")
	request, err := http.NewRequest(http.MethodGet, "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/registration/registerPostSignIn?mode=registration", nil)
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return UnknownHTTPResponseStatus
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}
	reader := strings.NewReader(string(body))
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return UnableToGetDocument
	}
	samlRequestValue := ""
	document.Find("input[name='SAMLRequest']").Each(func(index int, element *goquery.Selection) {
		value, exists := element.Attr("value")
		if exists {
			samlRequestValue = value
		}
	})

	s.SAMLRequest = samlRequestValue
	return nil
}

func (s *SignupTask) SubmitSamIsso() error {
	fmt.Println("Submitting Sam Isso")

	values := url.Values{
		"SAMLRequest": {s.SAMLRequest},
	}

	request, err := http.NewRequest(http.MethodPost, "https://eis-prod.ec.fhda.edu/samlsso", bytes.NewBufferString(values.Encode()))
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return UnknownHTTPResponseStatus
	}

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

	s.SAMLResponse = samlResponseValue
	return nil
}

func (s *SignupTask) SubmitSSBSp() error {
	fmt.Println("Submitting SSB Sp")

	values := url.Values{
		"SAMLResponse": {s.SAMLResponse},
	}

	request, err := http.NewRequest(http.MethodPost, "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/saml/SSO/alias/registrationssb-prod-sp", bytes.NewBufferString(values.Encode()))
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return UnknownHTTPResponseStatus
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}
	if len(body) > 0 {

	}
	return nil
}

func (s *SignupTask) SaveTerm() error {
	fmt.Println("Saving Term")

	url := fmt.Sprintf(
		"https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/term/saveTerm?mode=registration&term=%s",
		s.task.TermId,
	)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}

	if len(body) > 0 {
	}
	return nil
}

func (s *SignupTask) GetRegistrationStatus() error {
	fmt.Println("Getting registration status")

	termData := fmt.Sprintf("term=%s&studyPath=&studyPathText=&startDatepicker=&endDatepicker=&uniqueSessionId=", s.task.TermId)
	request, err := http.NewRequest(http.MethodPost, "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/term/search?mode=registration", bytes.NewBufferString(termData))
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return UnknownHTTPResponseStatus
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}

	registrationStatus := RegistrationStatus{}
	if err := json.Unmarshal(body, &registrationStatus); err != nil {
		log.Println(err)
		return UnableToParseJSON
	}
	if len(registrationStatus.StudentEligFailures) > 0 {
		var hasRegistrationTime bool
		var failure string
		for _, _failure := range registrationStatus.StudentEligFailures {
			if strings.Contains(_failure, "You can register from") {
				hasRegistrationTime = true
				failure = _failure
			}
		}

		if hasRegistrationTime {
			regex := regexp.MustCompile(`\d{2}/\d{2}/\d{4} \d{2}:\d{2} [APM]{2}`)
			matches := regex.FindAllString(failure, -1)

			if len(matches) > 0 {
				loc, _ := time.LoadLocation("America/Los_Angeles")
				targetTime, err := time.ParseInLocation("01/02/2006 03:04 PM", matches[0], loc)
				if err != nil {
					return FailedParsingDate
				}

				now := time.Now().In(loc)
				fmt.Printf("Registration opens at: %s\n", targetTime)
				if now.Before(targetTime) {

					timeToWait := targetTime.Sub(now) + 5*time.Second

					resumeDate := now.Add(timeToWait)
					fmt.Printf("Will continue after: %s\n", resumeDate.Format("2006-01-02 03:04:05 -0700 MST"))
					fmt.Printf("Waiting %s to continue\n", formatDuration(timeToWait))
					time.Sleep(timeToWait)
				} else {
				}
			}
		} else {
			return NotEligibleToRegister
		}
	}
	return nil
}

func (s *SignupTask) VisitClassRegistration() error {
	fmt.Println("Visiting class registration")

	request, err := http.NewRequest(http.MethodHead, "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/classRegistration/classRegistration", nil)
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
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

func (s *SignupTask) GetEvents() error {
	fmt.Println("Getting events")

	request, err := http.NewRequest(http.MethodGet, "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/classRegistration/getRegistrationEvents?termFilter=null", nil)
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return UnknownHTTPResponseStatus
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}

	if len(body) > 0 {
	}
	return nil
}

func (s *SignupTask) AddCourse(CourseNumber string) error {
	fmt.Println("Adding course")

	url := fmt.Sprintf(
		"https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/classRegistration/addRegistrationItem?term=%s&courseReferenceNumber=%s&olr=false",
		s.task.TermId, CourseNumber,
	)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return UnknownHTTPResponseStatus
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}

	addCourse := AddCourse{}
	if err := json.Unmarshal(body, &addCourse); err != nil {
		log.Println(err)
		return UnableToParseJSON
	}
	if addCourse.Success {
		dataModel, err := extractModel([]byte(body))
		if err != nil {
			log.Println(err)
			return UnableToParseJSON
		}
		s.Model = dataModel
	} else {
		if len(addCourse.Message) > 0 {
			fmt.Printf("Error adding course: %s\n", addCourse.Message)
		}
		return FailedToAddCourse
	}
	return nil
}

func (s *SignupTask) AddCourses() error {
	fmt.Println("Adding courses")

	for _, course := range s.task.CoursesToAdd {
		s.AddCourse(course)
	}
	return nil
}

func (s *SignupTask) SubmitChanges() error {
	fmt.Println("Submitting changes")

	batch := BatchUpdate{
		Update: []map[string]interface{}{s.Model},
	}

	payloadJson, err := json.MarshalIndent(batch, "", "  ")
	if err != nil {
		fmt.Println(err)
		return UnableToParseJSON
	}

	request, err := http.NewRequest(http.MethodPost, "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/classRegistration/submitRegistration/batch", bytes.NewBufferString(string(payloadJson)))
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Add("accept", "*/*")
	request.Header.Add("accept-language", "en-US,en;q=0.9")
	request.Header.Add("content-type", "application/json")
	request.Header.Add("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return UnknownHTTPResponseStatus
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}
	changes := Changes{}
	if err := json.Unmarshal(body, &changes); err != nil {
		fmt.Println(err)
		return UnableToParseJSON
	}

	for _, data := range changes.Data.Update {
		for _, course := range s.task.CoursesToAdd {
			if data.CourseReferenceNumber == course {
				if len(data.CrnErrors) > 0 || data.StatusDescription == "Errors Preventing Registration" {
					fmt.Printf("%d Errors encountered while adding %s - %s\n", len(data.CrnErrors), data.CourseReferenceNumber, data.CourseTitle)
					for _, error := range data.CrnErrors {
						fmt.Printf("Error received: %s\n", error.Message)
					}
				}

				if data.StatusDescription == "Registered" {
					fmt.Printf("Successfully registered for %s - %s\n", data.CourseReferenceNumber, data.CourseTitle)
					s.task.sendSuccessfulEnrollmentNotification(data.CourseTitle)
					return nil
				}
			}
		}
	}
	return nil
}

func (s *SignupTask) Run() error {

	steps := []func() error{
		s.VisitHomepage,
		s.Login,
		s.SubmitCommonAuth,
		s.SubmitSSOManager,
		s.RegisterPostSignIn,
		s.SubmitSamIsso,
		s.SubmitSSBSp,
		s.SaveTerm,
		s.GetRegistrationStatus,
		s.VisitClassRegistration,
		s.GetEvents,
		s.AddCourses,
		s.SubmitChanges,
	}

	for _, step := range steps {
		if err := Retry(s.task.RetryAmount, s.task.RetryDuration, step); err != nil {
			return MaximumAttemptsRetry
		}
	}

	s.task.Client.CloseIdleConnections()
	return nil
}

func NewSignupTask(task *Task) *SignupTask {
	return &SignupTask{task: task}
}
