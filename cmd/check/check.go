package main

import (
	"encoding/json"
	"net/http"

	"github.com/b3cat/rutaxapi"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.New()
)

func main() {
	client := &http.Client{}

	taxAPI, err := rutaxapi.FromFile(client, "config.toml")
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Trying to get ticket ID")
	ticketID, err := taxAPI.GetTicketID("t=20200915T1518&s=1280.00&fn=9280440300539716&i=11442&fp=3147580442&n=1")

	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Ticket ID: %s", ticketID)
	log.Info("Trying to get ticket info")
	ticketInfo, err := taxAPI.GetTicketInfo(ticketID)

	if err != nil {
		log.Fatal(err)
	}

	prettyTicketInfo, _ := json.MarshalIndent(ticketInfo, "", "  ")
	log.Infof("Ticket Info: %s", prettyTicketInfo)
	log.Infof("Сумма операций в чеке %.2f Рублей", ticketInfo.Operation.Sum/100)
	log.Info("Попробуем обновить сессию сами")

	if err = taxAPI.RefreshSession(); err != nil {
		log.Fatal(err)
	}
}
