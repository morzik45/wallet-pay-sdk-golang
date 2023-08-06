package walletpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
)

const (
	EventTypesOrderPaid   = "ORDER_PAID"
	EventTypesOrderFailed = "ORDER_FAILED"
)

// Event object structure (see https://docs.wallet.pay.finance/docs/webhooks)
type Event struct {
	EventDateTime string  `json:"eventDateTime"` // ISO-8601 date time when some event triggered this webhook message
	EventID       int64   `json:"eventId"`       // Idempotency key, for single event we send no more than 1 type of webhook message
	Type          string  `json:"type"`          // Type of payload. Currently, ORDER_PAID / ORDER_FAILED
	Payload       Payload `json:"payload"`       // Json payload of message, see "Payload object structure" below
}

// Payload object structure
type Payload struct {
	Status                 string              `json:"status"`                 // Order status, clarifying reason of FAIL (e.g. status=EXPIRED) Sent if type=ORDER_FAILED
	ID                     int64               `json:"id"`                     // Order id
	Number                 string              `json:"number"`                 // Human-readable (short) order number
	ExternalID             string              `json:"externalId"`             // Order ID in the Merchant system
	CustomData             string              `json:"customData"`             // Custom string given during order creation
	OrderAmount            MoneyAmount         `json:"orderAmount"`            // Order amount and currency code Format: { "currencyCode": "TON", "amount": "30.45" }
	SelectedPaymentOption  SelectPaymentOption `json:"selectedPaymentOption"`  // User selected payment option. Format: {"amount": {"currencyCode": "TON","amount": "10.0"},"exchangeRate": "1.0"} Sent if type=ORDER_PAID
	OrderCompletedDateTime string              `json:"orderCompletedDateTime"` // ISO-8601 date time when the order was PAID/FAILED
}

// SelectPaymentOption object structure
type SelectPaymentOption struct {
	Amount       MoneyAmount `json:"amount"`
	AmountFee    MoneyAmount `json:"amountFee"`
	AmountNet    MoneyAmount `json:"amountNet"`
	ExchangeRate string      `json:"exchangeRate"`
}

// ParseEvents parses a WalletPay webhook events
func ParseEvents(body []byte) ([]Event, error) {
	var events []Event
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (c *Client) computeSignature(
	httpMethod string,
	uriPath string,
	timestamp string,
	body string,
) string {
	base64Body := base64.StdEncoding.EncodeToString([]byte(body))
	stringToSign := httpMethod + "." + uriPath + "." + timestamp + "." + base64Body
	mac := hmac.New(sha256.New, []byte(c.ApiKey))
	mac.Write([]byte(stringToSign))
	byteArraySignature := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(byteArraySignature)
}

func (c *Client) VerifySignature(
	httpMethod string,
	uriPath string,
	timestamp string,
	body string,
	signature string,
) bool {
	return c.computeSignature(httpMethod, uriPath, timestamp, body) == signature
}
