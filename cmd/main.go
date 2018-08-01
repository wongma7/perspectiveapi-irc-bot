package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	irc "github.com/fluffle/goirc/client"
	"github.com/wongma7/perspectiveapi-irc-bot/pkg/toxic/perspective"
)

var lastMessages map[string]map[string]string

var (
	channels string
	server   string
)

func init() {
	flag.StringVar(&channels, "channels", "#neopets", "comma-separated list of channels")
	flag.StringVar(&server, "server", "irc.freenode.net", "e.g. irc.freenode.net")
}

func main() {
	flag.Parse()

	// Creating a simple IRC client is simple.
	c := irc.SimpleClient("HRBot")

	lastMessages := make(map[string]map[string]string)

	// Add handlers to do things here!
	// e.g. join a channel on connect.
	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			for _, channel := range strings.Split(channels, ",") {
				conn.Join(channel)
				lastMessages[channel] = make(map[string]string)
			}
		})
	// And a signal on disconnect
	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	// ... or, use ConnectTo instead.
	if err := c.ConnectTo(server); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}

	p := &perspective.Perspective{}
	c.HandleFunc(irc.PRIVMSG,
		func(conn *irc.Conn, line *irc.Line) {
			text := line.Text()
			words := strings.Fields(text)
			command := words[0]
			if command == "!hr" {
				if len(words) < 2 {
					return
				}

				culprit := words[1]

				comment := ""
				if len(words) == 2 {
					if message, ok := lastMessages[line.Target()][line.Nick]; ok {
						comment = message
					}
				} else {
					comment = strings.Replace(text, command, "", 1)
					comment = strings.Replace(comment, culprit, "", 1)
				}
				comment = strings.TrimSpace(comment)
				if comment == "" {
					return
				}
				fmt.Printf("Scoring comment %s by %s\n", comment, culprit)

				score, err := p.ScoreComment(comment)
				if err != nil {
					fmt.Printf("Error scoring comment: %s\n", err.Error())
					return
				}

				toxicity := strconv.FormatFloat(score*100, 'f', 1, 64) + "%"
				if len(comment) > 12 {
					comment = comment[0:12] + "..."
				}
				analysis := fmt.Sprintf("%s's message \"%s\" has a %s chance of being toxic.", culprit, comment, toxicity)
				c.Privmsg(line.Target(), analysis)
			} else {
				lastMessages[line.Target()][line.Nick] = text
			}
		})

	// Wait for disconnect
	<-quit
}
