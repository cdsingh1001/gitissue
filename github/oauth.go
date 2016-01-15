package github

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var AuthURL = "https://api.github.com/authorizations"

// Token type
// Contains token as a string and file where it is stored
type Token struct {
	key  string
	file string
}

// OAuth is an authentication structure exchanged with github.com
// in a JSON format
type OAuth struct {
	Scopes []string `json:"scopes"`
	Note   string   `json:"note"`
	Token  string   `json:"token"`
}

type Error struct {
	Resource string
	Field    string
	Code     string
}

// ErrorMsg contains the details of the error received from github.com
// when a request containing JSON data fails.
// ErrorMsg gives details on what failed - which field was missing or
// what went wrong with the JSON request
type ErrorMsg struct {
	Message string `json:"message"`
	Errors  []*Error
}

func (e ErrorMsg) String() string {
	result := fmt.Sprintf("Message: %s \n", e.Message)
	for _, v := range e.Errors {
		result += fmt.Sprintf("Resource: %s\n", v.Resource)
		result += fmt.Sprintf("Field: %s\n", v.Field)
		result += fmt.Sprintf("Code: %s\n", v.Code)
	}

	return result
}

// ReadToken reads the token string from a given file name
func ReadToken(file string) (string, error) {
	var key string
	fd, err := os.Open(file)
	if err != nil {
		return "", err
	}

	input := bufio.NewScanner(fd)
	if input.Scan() {
		key = input.Text()
	}
	fd.Close()
	return key, nil
}

func (t Token) Store(key, file string) error {
	fd, err := os.Create(file)
	if err != nil {
		return err
	}
	t.key = key
	t.file = file
	_, err = fd.WriteString(key)
	if err != nil {
		return err
	}
	fd.Close()

	return nil
}

func (t Token) Retrieve() (string, error) {
	var key string
	fd, err := os.Open(t.file)
	if err != nil {
		return "", err
	}

	input := bufio.NewScanner(fd)
	if input.Scan() {
		key = input.Text()
	}
	fd.Close()
	return key, nil
}

// GetOAuthToken is an API that send the request to github.com to
// create a new token on behalf of an application
// returns - a newly created token is returned
// this token is as good as password - use with care
func GetOAuthToken(user, pass, note string) (string, error) {
	var oauth OAuth
	token := ""

	oauth.Scopes = []string{"repo", "user"}
	oauth.Note = note

	data, err := json.MarshalIndent(oauth, "", "  ")
	if err != nil {
		return token, err
	}
	b := bytes.NewBuffer(data)

	client := &http.Client{}

	url := AuthURL

	req, err := http.NewRequest("POST", url, b)
	req.SetBasicAuth(user, pass)

	resp, err := client.Do(req)
	if err != nil {
		return token, err
	}

	if resp.StatusCode != http.StatusCreated {
		var em ErrorMsg
		if err := json.NewDecoder(resp.Body).Decode(&em); err != nil {
			return token, err
		}
		fmt.Println(em)
		return token, fmt.Errorf("POST failed: %s", resp.Status)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&oauth); err != nil {
		return token, err
	}

	token = oauth.Token

	return token, nil
}
