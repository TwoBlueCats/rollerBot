package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/TwoBlueCats/diceRolls"
	"go.uber.org/multierr"
	tele "gopkg.in/telebot.v3"
)

func sendToAdmins(b *tele.Bot, admins []int64, what interface{}, opts ...interface{}) ([]*tele.Message, error) {
	var mErr error
	var messages []*tele.Message
	for _, id := range admins {
		msg, err := b.Send(tele.ChatID(id), what)
		messages = append(messages, msg)
		if err != nil {
			mErr = multierr.Append(mErr, err)
		}
	}
	return messages, mErr
}

func main() {
	rand.Seed(time.Now().Unix())

	pref := tele.Settings{
		Token:   os.Getenv("BOT_TOKEN"),
		Poller:  &tele.LongPoller{Timeout: 10 * time.Second},
		Verbose: os.Getenv("DEBUG") != "",
	}

	adminsEnv := os.Getenv("BOT_ADMINS")
	if len(adminsEnv) == 0 {
		log.Fatal("no bot admins")
		return
	}
	admins := make([]int64, 0)
	for _, id := range strings.Split(adminsEnv, ",") {
		value, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Fatal(err)
			return
		}
		admins = append(admins, value)
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Print("Init: ok")

	_, err = sendToAdmins(b, admins, "I started")
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/hello", func(c tele.Context) error {
		return c.Send("Hello, " + c.Message().Sender.Username + "!")
	})

	b.Handle("/me", func(c tele.Context) error {
		sender := c.Message().Sender
		message := "Hello, @" + sender.Username + "\\!\n"
		message += "Your id is `" + c.Message().Sender.Recipient() + "`\n"
		return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
	})

	b.Handle("/report", func(c tele.Context) error {
		_, err := sendToAdmins(b, admins, "New report:\n"+c.Message().Payload)
		if err != nil {
			return c.Send("Please try again")
		}
		return c.Send("Thank for your report")
	})

	b.Handle("/roll", func(c tele.Context) error {
		res, err := diceRolls.Parser(c.Message().Payload)
		if err != nil {
			return err
		}

		value := res.Value()

		err = c.Send(strconv.Itoa(value))
		if err != nil {
			return err
		}
		err = c.Send(res.Description(false))
		if err != nil {
			return err
		}
		return c.Send(strconv.Itoa(value) + " = " + res.Description(true))
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		res, err := diceRolls.Parser(c.Text())
		if err != nil {
			return err
		}

		value := res.Value()

		err = c.Send(strconv.Itoa(value))
		if err != nil {
			return err
		}
		err = c.Send(res.Description(false))
		if err != nil {
			return err
		}
		return c.Send(strconv.Itoa(value) + " = " + res.Description(true))
	})

	b.Handle(tele.OnQuery, func(c tele.Context) error {
		res, err := diceRolls.Parser(c.Query().Text)
		if err != nil {
			return err
		}

		value := res.Value()

		results := make(tele.Results, 0)
		results = append(results, &tele.ArticleResult{
			Title:       "result",
			Text:        strconv.Itoa(value),
			Description: strconv.Itoa(value),
		})
		results = append(results, &tele.ArticleResult{
			Title:       "description",
			Text:        res.Description(false),
			Description: res.Description(false),
		})
		results = append(results, &tele.ArticleResult{
			Title:       "explain",
			Text:        strconv.Itoa(value) + " = " + res.Description(true),
			Description: strconv.Itoa(value) + " = " + res.Description(true),
		})
		for idx := range results {
			results[idx].SetResultID(strconv.Itoa(idx))
		}
		return c.Answer(&tele.QueryResponse{
			Results:   results,
			CacheTime: 60,
		})
	})

	b.Start()
}
