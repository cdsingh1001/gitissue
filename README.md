## gitissue
gitissue is command line tool to create/edit/search issues on github repositories

### Installation

Note: You will need [Go 1.5](https://golang.org/dl/) or newer.

1. go get github.com/cdsingh1001/gitissue
  
  (For help on go get utility, please see this: https://golang.org/cmd/go/)

2. Run gitissue (assumes $GOPATH/bin is in your $PATH)


### Usage:
To start using the tool, it is recommended to generate OAuth Token (need to do only once) for your github account

##### Generate OAuth Token:
gitissue oauth <note (reminder) for this OAuth Token>

##### Create a github issue:
    gitissue create

    -r repo

    -t "Title of the issue"

    -b "Body of the issue" -

##### Delete a github issue:
    gitissue edit

    -r repository (user/repo)

    -i issue number

    -s state of issue ("closed" to delete the issue)

##### Get a single github issue:
    gitissue get

    -r repository (user/repo)

    -i issue number

##### Search Github issues (from a repo):
    gitissue search

    -r repository (user/repo)

    -f filter
