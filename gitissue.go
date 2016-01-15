// gitissue is a tool to create/search/edit issues on github.com
package main

import (
	"flag"
	"fmt"
	"github.com/cdsingh1001/gitissue/github"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"syscall"
)

const usage = `gitissue is a command line tool to create/search/edit issues on github.com

Usage:
To start using the tool, it is recommended to generate OAuth Token (need to do only once)
for your github account

Generate OAuth Token:
gitissue oauth  <note (reminder) for this OAuth Token>

Create a github issue:
gitissue create 
    -r repo
    -t "Title of the issue"
    -b "Body of the issue"
-

Edit a github issue:
gitissue edit  
    -r repository (user/repo)
    -i issue number
    -s state of issue ("closed" to delete the issue)

Get a single github issue:
gitissue get
    -r repository (user/repo)
    -i issue number

Search Github issues (from a repo):
gitissue search
    issues search
    -r repository (user/repo)
    -f filter
...`

// Container to collect user input
type Input struct {
	user   string
	repo   string
	title  string
	body   string
	state  string
	label  string
	number int
	filter string
}

// Print user input in a customized format
func (i Input) String() string {
	result := fmt.Sprintf("User: %s\n", i.user)
	result += fmt.Sprintf("Repo: %s\n", i.repo)
	result += fmt.Sprintf("Title: %s\n", i.title)
	result += fmt.Sprintf("Body: %s\n", i.body)
	result += fmt.Sprintf("Issue number: %d\n", i.number)
	result += fmt.Sprintf("State: %s\n", i.state)
	result += fmt.Sprintf("Label: %s\n", i.label)
	result += fmt.Sprintf("Filter: %s\n", i.filter)
	return result
}

var input Input

type subCommand struct {
	flags *flag.FlagSet
}

func (s subCommand) Parse(args []string) error {
	s.flags.Parse(args)
	if s.flags.Parsed() != true {
		fmt.Println("Flags not parsed")
		PrintUsageAndExit(-1)
	}
	return nil
}

var create, edit, get, search subCommand

func getCredentials() (string, string) {

	var user string

	fmt.Printf("Enter Username: ")
	fmt.Scanf("%s", &user)
	fmt.Print("Enter Password: \n")
	bytePassword, _ := terminal.ReadPassword(syscall.Stdin)

	return user, string(bytePassword)
}

func Authenticate() error {
	user, pass := getCredentials()
	key, err := github.GetOAuthToken(user, pass, os.Args[2])
	if err != nil {
		log.Fatal(err)
		return err
	}

	t := new(github.Token)
	t.Store(key, "token")
	fmt.Println("Token generated successfully in file named token")
	return nil
}

func init() {
	create.flags = flag.NewFlagSet("create", flag.ExitOnError)
	create.flags.StringVar(&input.user, "u", "", "Username")
	create.flags.StringVar(&input.repo, "r", "", "Repository")
	create.flags.StringVar(&input.title, "t", "", "Title of the issue")
	create.flags.StringVar(&input.body, "b", "", "Body of the issue")

	edit.flags = flag.NewFlagSet("edit", flag.ExitOnError)
	edit.flags.StringVar(&input.user, "u", "", "Username")
	edit.flags.StringVar(&input.repo, "r", "", "Repository")
	edit.flags.StringVar(&input.state, "s", "", "State of the issue")
	edit.flags.StringVar(&input.label, "l", "", "Label")
	edit.flags.IntVar(&input.number, "i", 1, "Issue number")

	get.flags = flag.NewFlagSet("get", flag.ExitOnError)
	get.flags.IntVar(&input.number, "i", 1, "Issue number")
	get.flags.StringVar(&input.user, "u", "", "Username")
	get.flags.StringVar(&input.repo, "r", "", "Repository")

	search.flags = flag.NewFlagSet("find", flag.ExitOnError)
	search.flags.StringVar(&input.user, "u", "", "Username")
	search.flags.StringVar(&input.repo, "r", "", "Repository")
	search.flags.StringVar(&input.filter, "f", "", "Filter")
}

func main() {
	if len(os.Args) < 3 {
		PrintUsageAndExit(-1)
	}

	if os.Args[1] == "oauth" {
		Authenticate()
		os.Exit(0)
	}

	token, err := github.ReadToken("token")
	if err != nil {
		log.Fatal(err, " Please authenticate first\n")
	}

	switch os.Args[1] {
	case "create":
		create.Parse(os.Args[2:])
		user := github.UserInfo{input.user, token, input.repo}
		issue := github.Issue{Title: input.title, Body: input.body}
		newIssue, err := github.CreateIssue(&user, &issue)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("New issue created at %q\n", string(newIssue[0]))
	case "edit":
		edit.Parse(os.Args[2:])
		user := github.UserInfo{input.user, token, input.repo}
		issue := github.Issue{Number: input.number, State: input.state, Label: input.label}
		err = github.EditIssue(&user, &issue)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Issue edited successfully\n")
	case "get":
		get.Parse(os.Args[2:])
		user := github.UserInfo{input.user, token, input.repo}
		issue, err := github.GetIssue(&user, input.number)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(issue)
	case "search":
		search.Parse(os.Args[2:])
		user := github.UserInfo{input.user, token, input.repo}
		result, err := github.SearchIssues(&user, input.filter)
		if err != nil {
			log.Fatal(err)
		}
		for _, item := range result.Items {
			fmt.Println("#", item.Number,
				item.User.Login, item.Title)
		}
		fmt.Printf("%d issues: items returned %d\n", result.TotalCount, len(result.Items))
	default:
		PrintUsageAndExit(1)
	}

}

func PrintUsageAndExit(code int) {
	fmt.Printf("%s\n", usage)
	os.Exit(code)
}
