package rutaxapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
)

var (
	// APIBase ...
	APIBase     = "https://irkkt-mobile.nalog.ru:8888/v2/"
	deviceOS    = "Android"
	deviceID    = "1234"
	contentType = "application/json"
)

// Credentials ...
type Credentials struct {
	Session      string `toml:"session"`
	Secret       string `toml:"client_secret"`
	RefreshToken string `toml:"refresh_token"`
}

// TaxAPI ...
type TaxAPI struct {
	client                  *http.Client
	credChangeChan          chan Credentials
	credUpdatesCompleteChan chan bool
	creds                   *Credentials
}

// FromFile ...
func FromFile(client *http.Client, filePath string) (*TaxAPI, error) {
	creds := Credentials{}
	_, err := toml.DecodeFile(filePath, &creds)

	if err != nil {
		return nil, err
	}

	credChangeChan := make(chan Credentials)
	credUpdatesCompleteChan := make(chan bool)
	// Обновляем файл каждый раз
	go updateCreds(credChangeChan, credUpdatesCompleteChan, filePath)

	return &TaxAPI{
		credChangeChan:          credChangeChan,
		credUpdatesCompleteChan: credUpdatesCompleteChan,
		client:                  client,
		creds:                   &creds,
	}, nil
}

func updateCreds(credChangeChan chan Credentials, credUpdatesCompleteChan chan bool, output string) error {
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		return err
	}

	for {
		creds := <-credChangeChan
		encoder := toml.NewEncoder(file)

		err := encoder.Encode(creds)

		if err != nil {
			return err
		}

		credUpdatesCompleteChan <- true
	}
}

type refreshSessionRequestBody struct {
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshSessionResponseBody ...
type refreshSessionResponseBody struct {
	Session      string `json:"sessionId"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshSession ...
func (t *TaxAPI) RefreshSession() error {
	data := refreshSessionRequestBody{
		ClientSecret: t.creds.Secret,
		RefreshToken: t.creds.RefreshToken,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	endpoint := getEndpoint("mobile/users/refresh")
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Device-Id", deviceID)
	req.Header.Set("Device-OS", deviceOS)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("sessionId", t.creds.Session)

	res, err := t.client.Do(req)
	if err != nil {
		return err
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	result := new(refreshSessionResponseBody)
	err = json.Unmarshal(responseData, &result)
	if err != nil {
		return err
	}

	creds := Credentials{
		Session:      result.Session,
		RefreshToken: result.RefreshToken,
		Secret:       t.creds.Secret,
	}

	t.creds = &creds

	t.credChangeChan <- creds
	<-t.credUpdatesCompleteChan
	return nil
}

type getTicketRequestBody struct {
	Qr string `json:"qr"`
}

type getTicketResponseBody struct {
	ID string `json:"id"`
}

// GetTicketID ...
func (t *TaxAPI) GetTicketID(qrString string) (string, error) {
	data := getTicketRequestBody{
		Qr: qrString,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return "", nil
	}

	bodyBuf := bytes.NewBuffer(body)
	req, err := http.NewRequest("POST", getEndpoint("ticket"), bodyBuf)
	if err != nil {
		return "", nil
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", strconv.Itoa(bodyBuf.Len()))
	req.Header.Set("sessionId", t.creds.Session)

	res, err := t.client.Do(req)
	if err != nil {
		return "", err
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	result := new(getTicketResponseBody)
	err = json.Unmarshal(responseData, &result)

	if err != nil {
		return "", err
	}

	return result.ID, nil
}

// GetTicketInfoResponseBody ...
type GetTicketInfoResponseBody struct {
	ID        string  `json:"id"`
	Status    float64 `json:"status"`
	Operation struct {
		Date string  `json:"date"`
		Type float64 `json:"type"`
		Sum  float64 `json:"sum"`
	} `json:"operation"`
	Seller struct {
		Inn string `json:"inn"`
	} `json:"seller"`
}

// GetTicketInfo ...
func (t *TaxAPI) GetTicketInfo(ticketID string) (result GetTicketInfoResponseBody, err error) {
	url := getEndpoint(fmt.Sprintf("tickets/%s", ticketID))
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return result, nil
	}
	req.Header.Set("Device-OS", deviceOS)
	req.Header.Set("Device-Id", deviceID)
	req.Header.Set("sessionId", t.creds.Session)

	res, err := t.client.Do(req)
	if err != nil {
		return result, err
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(responseData, &result)

	if err != nil {
		return result, err
	}

	return result, nil
}

func getEndpoint(method string) string {
	return fmt.Sprintf("%s%s", APIBase, method)
}
