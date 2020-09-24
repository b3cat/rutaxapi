package rutaxapi

import (
	"image"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// GetTicketInfoByQr ...
func (t *TaxAPI) GetTicketInfoByQr(qr image.Image) (result TicketInfo, err error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(qr)
	if err != nil {
		return
	}
	// decode image
	qrReader := qrcode.NewQRCodeReader()
	res, err := qrReader.Decode(bmp, nil)
	if err != nil {
		return
	}
	text := res.GetText()

	t.log.Infof("QR Query: %s", text)
	ticketID, err := t.GetTicketID(text)
	if err != nil {
		return
	}

	result, err = t.GetTicketInfo(ticketID)
	return
}
