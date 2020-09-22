package rutaxapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
)

var (
	// APIBase ...
	APIBase = "https://irkkt-mobile.nalog.ru:8888/v2/"
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
	creds                   *Credentials
	credChangeChan          chan Credentials
	credUpdatesCompleteChan chan bool
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

	defer file.Close()

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

func (t *TaxAPI) getBaseRequestHeaders() http.Header {
	headersMap := map[string]string{
		"Sessionid":    t.creds.Session,
		"Device-OS":    "Android",
		"Device-ID":    "1234",
		"Content-Type": "application/json",
	}

	headers := http.Header{}

	for k, value := range headersMap {
		headers.Add(k, value)
	}

	return headers
}

func getEndpoint(method string) string {
	return fmt.Sprintf("%s%s", APIBase, method)
}

func (t *TaxAPI) makeRequest(method string, uri string, body []byte, extraHeaders map[string]string) (result []byte, err error) {
	endpoint := getEndpoint(uri)
	bodyBuf := bytes.NewBuffer(body)
	req, err := http.NewRequest(method, endpoint, bodyBuf)
	if err != nil {
		return
	}

	headers := t.getBaseRequestHeaders()
	for k, v := range extraHeaders {
		headers.Add(k, v)
	}
	req.Header = headers

	res, err := t.client.Do(req)

	if err != nil {
		return
	}

	// Протухла сессия
	if res.StatusCode == 498 {
		err = t.RefreshSession()
		if err != nil {
			return
		}

		// Попробуем сделать запрос по новой
		return t.makeRequest(method, uri, body, extraHeaders)
	}

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		err = fmt.Errorf("not success status code %d %s", res.StatusCode, res.Status)
		return
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	return responseData, nil
}
