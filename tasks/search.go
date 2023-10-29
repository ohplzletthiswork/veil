package tasks

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"time"

	http "github.com/bogdanfinn/fhttp"
)

type SearchTask struct {
	task       *Task
	courseInfo []CourseInfo
}

func (s *SearchTask) SearchForTerm() error {
	fmt.Println("Searching for term")

	data := url.Values{}
	data.Set("term", s.task.TermId)

	request, err := http.NewRequest(http.MethodPost, "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/term/search?mode=search", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return FailedToCreateRequest
	}

	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/x-www-form-urlencoded")
	request.Header.Set("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}

	fmt.Println(string(body))
	return nil
}

func (s *SearchTask) GetCourses() error {
	fmt.Println("Getting courses")

	url := fmt.Sprintf(
		"https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/searchResults/searchResults?txt_subject=%s&txt_term=%s&startDatepicker=&endDatepicker=&pageOffset=0&pageMaxSize=100&sortColumn=subjectDescription&sortDirection=asc",
		s.task.Subject, s.task.TermId,
	)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return FailedToCreateRequest
	}
	request.Header.Set("accept", "*/*")
	request.Header.Set("user-agent", s.task.UserAgent)

	resp, err := s.task.Client.Do(request)
	if err != nil {
		return FailedToMakeRequest
	}
	defer resp.Body.Close()

	readBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return FailedToReadResponseBody
	}

	coursesResponse := Courses{}
	if err := json.Unmarshal(readBytes, &coursesResponse); err != nil {
		fmt.Println(err)
		return UnableToParseJSON
	}

	if coursesResponse.Success == false {
		return CourseSearchUnsuccessful
	}

	fmt.Printf("Total count of courses: %d\n", coursesResponse.TotalCount)

	if coursesResponse.TotalCount == 0 {
		fmt.Println("No courses found")
		return CourseSearchUnsuccessful
	}

	var courses []CourseInfo
	for _, section := range coursesResponse.Data {
		for _, faculty := range section.Faculty {
			for _, meetingfaculty := range section.MeetingsFaculty {
				course := CourseInfo{
					TermDesc:              section.TermDesc,
					CourseReferenceNumber: faculty.CourseReferenceNumber,
					Subject:               section.Subject,
					CourseNumber:          section.CourseNumber,
					SequenceNumber:        section.SequenceNumber,
					CourseTitle:           section.CourseTitle,
					DisplayName:           faculty.DisplayName,
					BeginTime:             Convert24HourTimeTo12HourFormat(meetingfaculty.MeetingTime.BeginTime),
					EndTime:               Convert24HourTimeTo12HourFormat(meetingfaculty.MeetingTime.EndTime),
					StartDate:             meetingfaculty.MeetingTime.StartDate,
					EndDate:               meetingfaculty.MeetingTime.EndDate,
					MeetingType:           meetingfaculty.MeetingTime.MeetingTypeDescription,
					Room:                  meetingfaculty.MeetingTime.Room,
					MaximumEnrollment:     section.MaximumEnrollment,
					Enrollment:            section.Enrollment,
					SeatsAvailable:        section.SeatsAvailable,
					WaitAvailable:         section.WaitAvailable,
				}
				courses = append(courses, course)
			}
		}
	}
	s.courseInfo = courses
	return nil
}

func (s *SearchTask) ExportSearchData() error {

	fmt.Println("Exporting search data")

	currentTime := time.Now()
	fileName := fmt.Sprintf("%s.csv", currentTime.Format("2006-01-02_15-04-05"))
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Term", "Course Reference Number", "Subject", "Course Number", "Sequence Number", "Course Title", "Display Name", "Begin Time", "End Time", "Start Date", "End Date", "Meeting Type", "Room", "Maximum Enrollment", "Enrollment", "Seats Available", "Waitlist Available"}
	fmt.Printf("Writing to %s\n", fileName)
	err = writer.Write(header)
	if err != nil {
		return FailedToWrite
	}
	for _, course := range s.courseInfo {
		record := []string{
			course.TermDesc,
			course.CourseReferenceNumber,
			course.Subject,
			course.CourseNumber,
			course.SequenceNumber,
			course.CourseTitle,
			course.DisplayName,
			course.BeginTime,
			course.EndTime,
			course.StartDate,
			course.EndDate,
			course.MeetingType,
			course.Room,
			strconv.Itoa(course.MaximumEnrollment),
			strconv.Itoa(course.Enrollment),
			strconv.Itoa(course.SeatsAvailable),
			strconv.Itoa(course.WaitAvailable),
		}
		err = writer.Write(record)
		if err != nil {
			return FailedToWrite
		}
	}
	fmt.Println("Exported search data")
	return nil
}

func (s *SearchTask) Run() error {
	steps := []func() error{
		s.SearchForTerm,
		s.GetCourses,
		s.ExportSearchData,
	}

	for _, step := range steps {
		if err := Retry(s.task.RetryAmount, s.task.RetryDuration, step); err != nil {
			return MaximumAttemptsRetry
		}
	}

	s.task.Client.CloseIdleConnections()
	return nil
}

func NewSearchTask(task *Task) *SearchTask {
	return &SearchTask{task: task}
}
