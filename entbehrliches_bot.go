package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

var wikiurls = []string{
	"://de.wikipedia.org/wiki",
	"://de.m.wikipedia.org/wiki",
	"://en.wikipedia.org/wiki",
	"://en.m.wikipedia.org/wiki",
}

func containsWikiURL(msg string) bool {
	var contains bool = false
	for _, s := range wikiurls {
		if strings.Contains(msg, s) {
			contains = true
		}
	}
	return contains
}

func main() {

	// Arguments
	var (
		apiToken = flag.String("apitoken", "", "Telegram API Token")
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

	b.Handle(tb.OnText, func(m *tb.Message) {
		fmt.Println(m.Text)
		if containsWikiURL(m.Text) {
			b.Send(m.Chat, "Wiki url gefunden")
		}
	})

	b.Start()
}
