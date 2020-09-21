package main

import (
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/b3cat/rutaxapi"
	"github.com/sirupsen/logrus"
)

// Options ...
type Options struct {
	Secret       string `toml:"client_secret"`
	Session      string `toml:"session"`
	RefreshToken string `toml:"refresh_token"`
}

var (
	log = logrus.New()
)

func main() {
	options := Options{}
	_, err := toml.DecodeFile("config.toml", &options)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Config successfuly red")

	client := &http.Client{}

	taxAPI := rutaxapi.New(client, options.Session, options.Secret, options.RefreshToken)

	ticketID, err := taxAPI.GetTicketID("t=20200915T1518&s=1280.00&fn=9280440300539716&i=11442&fp=3147580442&n=1")

	if err != nil {
		log.Fatal(err)
	}

	ticketInfo, err := taxAPI.GetTicketInfo(ticketID)

	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Сумма операций в чеке %.2f Рублей", ticketInfo.Operation.Sum/100)
}
