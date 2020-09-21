package rutaxapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	// APIBase ...
	APIBase     = "https://irkkt-mobile.nalog.ru:8888/v2/"
	deviceOS    = "Android"
	deviceID    = "1234"
	contentType = "application/json"
)

// TaxAPI ...
type TaxAPI struct {
	client       *http.Client
	session      string
	secret       string
	refreshToken string
}

// New ...
func New(client *http.Client, session string, secret string, refreshToken string) *TaxAPI {
	return &TaxAPI{
		client:       client,
		session:      session,
		secret:       secret,
		refreshToken: refreshToken,
	}
}

type refreshSessionRequestBody struct {
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshSessionResponseBody ...
type refreshSessionResponseBody struct {
	Session      string `json:"session_id"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshSession ...
func (t *TaxAPI) RefreshSession() error {
	data := refreshSessionRequestBody{
		ClientSecret: t.secret,
		RefreshToken: t.refreshToken,
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
	req.Header.Set("sessionId", t.session)

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

	t.session = result.Session
	t.refreshToken = result.RefreshToken
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
		fmt.Printf("error at encoding body")
		return "", nil
	}

	bodyBuf := bytes.NewBuffer(body)
	req, err := http.NewRequest("POST", getEndpoint("ticket"), bodyBuf)
	if err != nil {
		return "", nil
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", strconv.Itoa(bodyBuf.Len()))
	req.Header.Set("sessionId", t.session)

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
	req.Header.Set("sessionId", t.session)

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
