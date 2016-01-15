// github provides API's to access api.github.com
// Using github package one can build an application
// to create/search/edit github issues
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var IssuesURL = "https://api.github.com/search/issues"
var UserURL = "https://api.github.com/users/"
var CreateURL = "https://api.github.com/"

// IssueSearchResult is a container for all the issues
// returned by a search query. JSON Format
type IssuesSearchResult struct {
	TotalCount int `json:"total_count"`
	Items      []*Issue
}

// Issue defines the format of a github issue
type Issue struct {
	Number    int
	HTMLURL   string `json:"html_url"`
	Title     string `json:"title"`
	State     string `json:"state"`
	User      *User
	Assignee  *Assignee
	CreatedAt time.Time `json:"created_at"`
	Body      string    `json:"body"`
	Label     string    `json:"label"`
}

type User struct {
	Login   string
	HTMLURL string `json:"html_url"`
}

type Assignee struct {
	Login string
}

type UserInfo struct {
	User  string
	Token string
	Repo  string
}

func (i Issue) String() string {
	result := fmt.Sprintf("Number: %d\n", i.Number)
	result += fmt.Sprintf("URL: %s\n", i.HTMLURL)
	result += fmt.Sprintf("Title: %s\n", i.Title)
	result += fmt.Sprintf("State: %s\n", i.State)
	result += fmt.Sprintf("Label: %s\n", i.Label)
	return result
}

func httpRequest(request, url, token string, b *bytes.Buffer) (*http.Response, error) {

	client := &http.Client{}

	req, err := http.NewRequest(request, url, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func verifyResp(resp *http.Response, expected int) *ErrorMsg {
	if resp.StatusCode != expected {
		var em ErrorMsg
		_ = json.NewDecoder(resp.Body).Decode(&em)
		return &em
	}
	return nil
}

func printResp(resp *http.Response) string {
	if resp != nil {
		return resp.Status
	} else {
		return ""
	}
}

// GetIssue  - given a single issues number -
// returns the details of a single issue
func GetIssue(u *UserInfo, number int) (*Issue, error) {

	issue := Issue{Number: number}
	data, _ := json.Marshal(&issue)
	b := bytes.NewBuffer(data)

	num := fmt.Sprintf("%d", number)
	request := "GET"
	url := CreateURL + "repos/" + u.Repo + "/issues/" + num
	token := u.Token
	resp, err := httpRequest(request, url, token, b)
	if err != nil {
		return nil, fmt.Errorf("%s failed: %s", request, printResp(resp))
	}
	defer resp.Body.Close()

	if errMsg := verifyResp(resp, http.StatusOK); errMsg != nil {
		fmt.Println(errMsg)
		return nil, fmt.Errorf("%s failed: %s\n", request, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return &issue, nil

}

// EditIssue - edits a single issue on github.com
// Take in issue details (just the fields that have changed)
// returns error (nil on success)
func EditIssue(u *UserInfo, i *Issue) error {

	issue, err := GetIssue(u, i.Number)
	if err != nil {
		return fmt.Errorf("failed to get issue: %d", i.Number)
	}

	i.Title = issue.Title

	data, _ := json.Marshal(&i)
	b := bytes.NewBuffer(data)

	num := fmt.Sprintf("%d", i.Number)
	request := "PATCH"
	url := CreateURL + "repos/" + u.Repo + "/issues/" + num
	token := u.Token

	resp, err := httpRequest(request, url, token, b)
	if err != nil {
		return fmt.Errorf("%s failed: %s", request, resp.Status)
	}
	defer resp.Body.Close()
	if errMsg := verifyResp(resp, http.StatusOK); errMsg != nil {
		fmt.Println(errMsg)
		return fmt.Errorf("%s failed: %s\n", request, resp.Status)
	}

	return nil
}

// CreateIssue - creates a single issue on github.com
// Take in issue details
// returns the URL of the newly created issue
func CreateIssue(u *UserInfo, i *Issue) ([]string, error) {
	var result = make([]string, 0)
	data, _ := json.Marshal(&i)
	b := bytes.NewBuffer(data)

	request := "POST"
	url := CreateURL + "repos/" + u.Repo + "/issues"
	token := u.Token

	resp, err := httpRequest(request, url, token, b)
	if err != nil {
		return nil, fmt.Errorf("%s failed: %s", request, resp.Status)
	}
	defer resp.Body.Close()
	if errMsg := verifyResp(resp, http.StatusCreated); errMsg != nil {
		fmt.Println(errMsg)
		return result, fmt.Errorf("%s failed: %s\n", request, resp.Status)
	}

	issueURL := resp.Header["Location"]

	result = issueURL

	return result, nil
}

// SearchIssue - searches for all the issues for a given repository
// with a given filter on github.com
// returns all the issues found matching the given repo/filter
func SearchIssues(u *UserInfo, filter string) (*IssuesSearchResult, error) {
	var result IssuesSearchResult

	terms := make([]string, 2)
	terms[0] = "repo:" + u.Repo
	if filter != "" {
		terms[1] = "is:" + filter
	}
	query := strings.Join(terms, " ")

	q := url.QueryEscape(query)

	IssuesURL = IssuesURL + "?q=" + q + "&page=1"
	fmt.Println(IssuesURL)
	client := &http.Client{}

	for IssuesURL != "" {
		req, err := http.NewRequest("GET", IssuesURL, nil)
		req.Header.Set("Authorization", "token "+u.Token)
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if errMsg := verifyResp(resp, http.StatusOK); errMsg != nil {
			fmt.Println(errMsg)
			return &result, fmt.Errorf("search query failed: %s", resp.Status)
		}

		link := resp.Header["Link"]
		if len(link) != 0 && strings.Contains(link[0], "next") {
			s := strings.Split(link[0], ";")
			nextURL := strings.TrimPrefix(s[0], "<")
			nextURL = strings.TrimSuffix(nextURL, ">")
			IssuesURL = nextURL
		} else {
			IssuesURL = ""
		}

		var tempResult IssuesSearchResult
		if err := json.NewDecoder(resp.Body).Decode(&tempResult); err != nil {
			return nil, err
		}
		result.TotalCount = tempResult.TotalCount
		for _, v := range tempResult.Items {
			result.Items = append(result.Items, v)
		}
	}

	return &result, nil
}
