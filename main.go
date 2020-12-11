package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	t "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

const (
	howto = `You need to say song name and artist name.
How to: "/search song name, artist name"`
	help = `Hey
Welcome to S-Lyrics Bot.
You can simply find your lyrics by using "/search" command.

` + howto
	tryanother = `
There is no lyrics to show 😔
We are sorry about that.

We are working on finding more lyrics.
Try another song 🙏
`
)

type lyric struct {
	Lines []string `json:"Lines"`
}

func init() {
	godotenv.Load(".env")
}

func main() {
	bot, err := t.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := t.NewUpdate(0)
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalln(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s]", update.Message.From.UserName)
		switch update.Message.Command() {
		case "help", "start":
			msg := t.NewMessage(update.Message.Chat.ID, help)
			bot.Send(msg)
		case "search":
			query := update.Message.CommandArguments()
			queries := strings.Split(query, ",")
			if len(queries) != 2 {
				msg := t.NewMessage(update.Message.Chat.ID, howto)
				bot.Send(msg)

				continue
			}

			songName := strings.TrimSpace(queries[0])
			artistName := strings.TrimSpace(queries[1])
			if artistName == "" || songName == "" {
				msg := t.NewMessage(update.Message.Chat.ID, howto)
				bot.Send(msg)

				continue
			}

			lines, err := getResultFromSite(artistName, songName)
			if err != nil {
				log.Println(err)
				msg := t.NewMessage(update.Message.Chat.ID, `The API Service is not available just for now. Try later.`)
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)

				continue
			}
			if len(lines) == 0 {
				msg := t.NewMessage(update.Message.Chat.ID, tryanother)
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)

				continue
			}

			msg := t.NewMessage(update.Message.Chat.ID, strings.Join(lines, "\n"))
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		default:
			msg := t.NewMessage(update.Message.Chat.ID, `Command is not found. Check "/help".`)
			bot.Send(msg)
		}
	}
}

func getResultFromSite(artistName, songName string) ([]string, error) {
	u, _ := url.Parse(os.Getenv("SEARCH_API_ADDRESS"))
	q, _ := url.ParseQuery(u.RawQuery)

	q.Add("artistName", artistName)
	q.Add("songName", songName)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("POST", fmt.Sprint(u), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("api-key", os.Getenv("SEARCH_API_KEY"))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		fmt.Println(res.StatusCode)
		return nil, errors.New("The API is not usable.")
	}

	l := new(lyric)
	err = json.NewDecoder(res.Body).Decode(&l)
	if err != nil {
		return nil, errors.New("Error occured while reading the json.")
	}

	return l.Lines, nil
}
