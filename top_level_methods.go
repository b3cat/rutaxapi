package rutaxapi

import (
	"image"

	"github.com/liyue201/goqr"
)

// GetTicketInfoByQr ...
func (t *TaxAPI) GetTicketInfoByQr(qr image.Image) (result TicketInfo, err error) {
	qrCodes, err := goqr.Recognize(qr)
	if err != nil {
		return
	}

	qrString := string(qrCodes[0].Payload)
	ticketID, err := t.GetTicketID(qrString)
	if err != nil {
		return
	}

	result, err = t.GetTicketInfo(ticketID)
	return
}
