package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyokomi/emoji/v2"

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
		orig_m   *tb.Message
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

	// Telegram Answering
	// This data format is insane...
	j := tb.ReplyButton{Text: emoji.Sprintf("Oh, ja! Toller Service hier!")}
	n := tb.ReplyButton{Text: "Nein, danke"}
	menu := [][]tb.ReplyButton{
		[]tb.ReplyButton{j},
		[]tb.ReplyButton{n},
	}

	// Handler for all messages
	log.Println("Starting Telegram Bot")
	b.Handle(tb.OnText, func(m *tb.Message) {

		// If message is an answer to a previous one
		if m.IsReply() && m.Text == j.Text && orig_m != nil {
			rxStrict := xurls.Strict()
			log.Printf("Ich mach mal nen GH ISSUE auf\n")
			log.Printf("%v", rxStrict.FindString(orig_m.Text))
			b.Send(m.Chat, emoji.Sprintf(":dog:Danke für deinen Beitrag, du toller Mensch :heart:"))
			orig_m = nil
		}

		var count int = 0
		url, containsurl := containsWikiURL(m.Text)
		if containsurl {

			log.Printf("Searching for %s in archives...\n", url)
			for _, s := range existing_urls {
				if strings.Contains(s, url) {
					count = count + 1
					log.Printf("Found %s in articles.\n", url)
					break
				}
			}

			if count > 0 {
				b.Send(m.Chat, emoji.Sprintf(":dog:*Jaul* Der Artikel kommt mir doch sehr bekannt vor, ich denke den hatten wir schon!"), tb.NoPreview, &tb.SendOptions{
					ReplyTo: m})
			} else {
				b.Send(m.Chat, emoji.Sprintf(":flushed:Wuff, den kenn ich garnicht! Willst du ihn vielleicht einreichen?"), &tb.SendOptions{ReplyTo: m}, &tb.ReplyMarkup{
					ReplyKeyboard:       menu,
					ForceReply:          true,
					ReplyKeyboardRemove: true,
					OneTimeKeyboard:     true,
				})
				// Remember message
				orig_m = m
			}
		}
	})

	b.Start()
}
