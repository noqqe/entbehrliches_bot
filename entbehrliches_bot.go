package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mvdan.cc/xurls/v2"

	tb "gopkg.in/tucnak/telebot.v2"
)

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
			return rxStrict.FindString(msg), true
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

func main() {

	// Arguments
	var (
		apiToken = flag.String("apitoken", "", "Telegram API Token")
		posts    = flag.String("posts", "", "Directory containing entbehrliches posts")
	)
	flag.Parse()

	// Validate Arguments
	if len(*apiToken) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	b, err := tb.NewBot(tb.Settings{
		Token:  *apiToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	// Find all files
	log.Println("Finding all markdown files")
	files := findMDFiles(*posts)

	// var existing_wikilinks []string
	var existing_urls []string

	log.Println("Loading all urls")
	for _, f := range files {
		//existing_wikilinks = append(existing_wikilinks, searchFileforWikiLink(f)[0])
		existing_urls = append(existing_urls, searchFileforWikiLink(f)...)
	}

	log.Println("Starting Telegram Bot")
	b.Handle(tb.OnText, func(m *tb.Message) {

		var count int = 0
		url, containsurl := containsWikiURL(m.Text)
		if containsurl {

			log.Printf("Searching for %s in archives...\n", url)
			for _, s := range existing_urls {
				if strings.Contains(s, url) {
					count = count + 1
					log.Printf("Found %s in articles.\n", url)
				}
			}

			if count > 0 {
				b.Send(m.Chat, "Hatten wa schon!")
			} else {
				b.Send(m.Chat, "Cool, kennsch garnet")
			}
		}
	})

	b.Start()
}
