package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"

	"github.com/b3cat/rutaxapi"
	"github.com/sirupsen/logrus"
)

var (
	log        = logrus.New()
	qrPath     = flag.String("qr", "qr.jpg", "jpeg file of qr code")
	configPath = flag.String("config", "config.toml", ".toml file config path")
)

func main() {
	flag.Parse()
	client := &http.Client{}

	taxAPI, err := rutaxapi.FromFile(client, *configPath)
	if err != nil {
		log.Fatal(err)
	}

	imgdata, err := ioutil.ReadFile(*qrPath)
	if err != nil {
		log.Fatal(err)
	}

	qr, _, err := image.Decode(bytes.NewReader(imgdata))
	if err != nil {
		log.Fatal(err)
	}

	ticketInfo, err := taxAPI.GetTicketInfoByQr(qr)
	if err != nil {
		log.Fatal(err)
	}

	prettyTicketInfo, _ := json.MarshalIndent(ticketInfo, "", "  ")
	log.Infof("Ticket Info: %s", prettyTicketInfo)

	log.Info("Попробуем обновить сессию сами")
	if err = taxAPI.RefreshSession(); err != nil {
		log.Fatal(err)
	}
}
