package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"net/http"

	"github.com/b3cat/rutaxapi"
	"github.com/liyue201/goqr"
	"github.com/sirupsen/logrus"
)

var (
	log        = logrus.New()
	qrPath     = flag.String("qr", "qr.jpg", "jpeg file of qr code")
	configPath = flag.String("config", "config.toml", ".toml file config path")
)

func recognizeQrCode(path string) (result string, err error) {
	imgdata, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	img, _, err := image.Decode(bytes.NewReader(imgdata))
	if err != nil {
		return
	}
	qrCodes, err := goqr.Recognize(img)
	if err != nil {
		return
	}

	return string(qrCodes[0].Payload), nil
}

func main() {
	flag.Parse()
	client := &http.Client{}

	taxAPI, err := rutaxapi.FromFile(client, *configPath)
	if err != nil {
		log.Fatal(err)
	}

	qr, err := recognizeQrCode(*qrPath)
	if err != nil {
		return
	}

	log.Info("Trying to get ticket ID")
	ticketID, err := taxAPI.GetTicketID(qr)

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
