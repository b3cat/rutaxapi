package rutaxapi

import (
	"encoding/json"
	"fmt"
)

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
func (t *TaxAPI) RefreshSession() (err error) {
	data := refreshSessionRequestBody{
		ClientSecret: t.creds.Secret,
		RefreshToken: t.creds.RefreshToken,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return
	}

	responseData, err := t.makeRequest("POST", "mobile/users/refresh", body, nil)
	if err != nil {
		return
	}

	result := new(refreshSessionResponseBody)
	err = json.Unmarshal(responseData, &result)
	if err != nil {
		return
	}

	creds := Credentials{
		Session:      result.Session,
		RefreshToken: result.RefreshToken,
		Secret:       t.creds.Secret,
	}

	t.creds = &creds

	t.credChangeChan <- creds
	<-t.credUpdatesCompleteChan
	return
}

type getTicketRequestBody struct {
	Qr string `json:"qr"`
}

type getTicketResponseBody struct {
	ID string `json:"id"`
}

// GetTicketID ...
func (t *TaxAPI) GetTicketID(qrString string) (id string, err error) {
	data := getTicketRequestBody{
		Qr: qrString,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return
	}

	responseData, err := t.makeRequest("POST", "ticket", body, nil)
	if err != nil {
		return
	}

	result := new(getTicketResponseBody)
	err = json.Unmarshal(responseData, &result)

	if err != nil {
		return
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
	uri := fmt.Sprintf("tickets/%s", ticketID)
	responseData, err := t.makeRequest("GET", uri, nil, nil)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(responseData, &result)

	if err != nil {
		return result, err
	}

	return result, nil
}
