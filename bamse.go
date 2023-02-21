package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v34/github"
	"github.com/kyokomi/emoji/v2"
	"golang.org/x/oauth2"
	tb "gopkg.in/tucnak/telebot.v2"
	"mvdan.cc/xurls/v2"
)

var Version = "unknown"

type IssueRequest struct {
	Title     *string   `json:"title,omitempty"`
	Body      *string   `json:"body,omitempty"`
	Labels    *[]string `json:"labels,omitempty"`
	Assignee  *string   `json:"assignee,omitempty"`
	State     *string   `json:"state,omitempty"`
	Milestone *int      `json:"milestone,omitempty"`
	Assignees *[]string `json:"assignees,omitempty"`
}

// Checks if message contains wiki url
func containsWikiURL(msg string) (string, bool) {

	var wikiurls = []string{
		"://de.wikipedia.org/wiki",
		"://de.m.wikipedia.org/wiki",
		"://en.wikipedia.org/wiki",
		"://en.m.wikipedia.org/wiki",
	}

	for _, s := range wikiurls {
		if strings.Contains(msg, s) {
			rxStrict := xurls.Strict()
			// Remove mobile urls to prevent duplication
			url_w_o_de := strings.Replace(rxStrict.FindString(msg), "://de.m.", "://de.", 1)
			url_w_o_en := strings.Replace(url_w_o_de, "://en.m.", "://en.", 1)
			url_obj, _ := url.Parse(url_w_o_en)
			url_w_o_query := url_obj.Scheme + "://" + url_obj.Host + url_obj.Path
			return url_w_o_query, true
		}
	}
	return "", false
}

// Generates a list of all markdown files
func findMDFiles(root string) []string {
	var matches []string

	d, err := os.Open(root)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ".md" {
				matches = append(matches, root+"/"+file.Name())
			}
		}
	}
	return matches
}

func existsInGithubIssues(url string) bool {
	var token string = os.Getenv("GITHUB_TOKEN")
	var repo_owner string = os.Getenv("GITHUB_OWNER")
	var repo_name string = os.Getenv("GITHUB_REPO")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	for page := 0; page < 10; page++ {
		ir := &github.IssueListByRepoOptions{ListOptions: github.ListOptions{PerPage: 100, Page: page}}
		i, _, _ := client.Issues.ListByRepo(ctx, repo_owner, repo_name, ir)
		for issue := 0; issue < len(i); issue++ {
			if strings.Contains(*i[issue].Title, url) {
				return true
			}
		}
	}
	return false
}

func createGithubIssue(via, url string) string {
	var token string = os.Getenv("GITHUB_TOKEN")
	var repo_owner string = os.Getenv("GITHUB_OWNER")
	var repo_name string = os.Getenv("GITHUB_REPO")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	t := fmt.Sprintf("Bitte hinzufügen %s", url)
	b := fmt.Sprintf("Der Link %s wurde uns von %s zugesendet", url, via)
	ir := &github.IssueRequest{
		Title: &t,
		Body:  &b,
	}

	i, _, _ := client.Issues.Create(ctx, repo_owner, repo_name, ir)
	return *i.HTMLURL
}

// Searches a single markdown file for wiki url occourences
// Returns slice [http.., http...]
func searchFileforWikiLink(f string) []string {
	fileBytes, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Println(err)
	}
	rxStrict := xurls.Strict()
	return rxStrict.FindAllString(string(fileBytes), -1)
}

// Build list of all posted wiki links so far
func initPosts(posts string) []string {

	// Find all files
	log.Println("Finding all markdown files")
	files := findMDFiles(posts)

	var existing_urls []string

	log.Println("Loading all urls")
	for _, f := range files {
		existing_urls = append(existing_urls, searchFileforWikiLink(f)...)
	}

	return existing_urls
}

func main() {

	// Arguments
	var (
		apiToken string = os.Getenv("APITOKEN")
		posts    string = os.Getenv("POSTS")
	)

	// Init new Telegram Bot
	b, err := tb.NewBot(tb.Settings{
		Token:  apiToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	// Init wiki list
	existing_urls := initPosts(posts)

	// Handler for all messages
	log.Printf("Starting Bamse Bot Version %s\n", Version)
	b.Handle(tb.OnText, func(m *tb.Message) {

		url, containsurl := containsWikiURL(m.Text)
		if containsurl {

			log.Printf("Searching for %s in archives...\n", url)
			var count int = 0
			for _, s := range existing_urls {
				if strings.Contains(s, url) {
					count = count + 1
					log.Printf("Found %s in articles.\n", url)
					break
				}
			}

			if existsInGithubIssues(url) {
				count = count + 1
				log.Printf("Found %s in Github Issues.\n", url)
			}

			if count > 0 {
				b.Send(m.Chat,
					emoji.Sprintf(":dog:Der Artikel kommt mir doch sehr bekannt vor, ich denke den hatten wir schon!"),
					tb.NoPreview,
					&tb.SendOptions{ReplyTo: m},
					&tb.ReplyMarkup{
						ReplyKeyboard:       [][]tb.ReplyButton{},
						ReplyKeyboardRemove: true,
					})
			} else {
				b.Send(m.Chat,
					emoji.Sprintf(":tada::tada::tada:Den Artikel kannte ich noch garnicht! Hab ihn eingereicht!"),
					tb.NoPreview,
					&tb.SendOptions{ReplyTo: m},
					&tb.ReplyMarkup{
						ReplyKeyboard:       [][]tb.ReplyButton{},
						ReplyKeyboardRemove: true,
					})
				rxStrict := xurls.Strict()
				createGithubIssue(m.Sender.Username, rxStrict.FindString(m.Text))
			}
		}
	})

	b.Start()
}
