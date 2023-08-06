package walletpay

import (
	"context"
	"fmt"
	"io"
	"net/url"
)

const (
	OrderStatusSuccess      = "SUCCESS"
	OrderStatusAlready      = "ALREADY"
	OrderStatusConflict     = "CONFLICT"
	OrderStatusAccessDenied = "ACCESS_DENIED"

	TON  = "TON"
	BTC  = "BTC"
	USDT = "USDT"
	EUR  = "EUR"
	USD  = "USD"
	RUB  = "RUB"
)

// MoneyAmount object structure
type MoneyAmount struct {
	CurrencyCode string `json:"currencyCode"` // Required. Enum: "TON" "BTC" "USDT" "EUR" "USD" "RUB" Currency code
	Amount       string `json:"amount"`       // Required. Big decimal string representation
}

type OrderRequest struct {
	Amount             MoneyAmount `json:"amount"`             // Required. Order mount
	Description        string      `json:"description"`        // Required. Description of order. Length is limited from 5 to 100 characters.
	ReturnURL          string      `json:"returnUrl"`          // Optional. URL to redirect after paying order, length is limited to 255 characters.
	FailReturnURL      string      `json:"failReturnUrl"`      // Optional. URL to redirect after unsuccessful order completion (expiration/cancellation/etc). Length is limited to 255 characters.
	CustomData         string      `json:"customData"`         // Optional. Any custom string, will be provided through webhook and order status polling, length is limited to 255 characters
	ExternalID         string      `json:"externalId"`         // Required. Order ID in the Merchant system. Use to prevent orders duplication due to request retries. Length is limited to 255 characters
	TimeoutSeconds     int64       `json:"timeoutSeconds"`     // Required. Order TTL, if the order is not paid within the timeout period. Min: 30, Max: 86400
	CustomerTelegramID int64       `json:"customerTelegramId"` // Required. The customer's telegram id (User_id).
}

type OrderResponse struct {
	// Status present Enum: "SUCCESS" "ALREADY" "CONFLICT" "ACCESS_DENIED"
	// SUCCESS - new order created; data is present.
	// ALREADY - order with completely same parameters including externalId already exists; data is present.
	// CONFLICT - order with different parameters but same externalId already exists; data is absent.
	// ACCESS_DENIED - you're not permitted to create orders, contact support in this case; data is absent.
	Status  string       `json:"status"`  // Required.
	Message string       `json:"message"` // Optional. Verbose reason of non-success result
	Preview OrderPreview `json:"data"`    // Optional. Order preview data. Present if status=SUCCESS
}

type OrderPreview struct {
	ID                 string      `json:"id"`                 // Required. Order id
	Status             string      `json:"status"`             // Required. Order status. Enum: "ACTIVE" "EXPIRED" "PAID" "CANCELLED"
	Number             string      `json:"number"`             // Required. Human-readable (short) order number
	Amount             MoneyAmount `json:"amount"`             // Required. Order amount
	CreateDateTime     string      `json:"createDateTime"`     // Required. ISO-8601 date time when the order was created
	ExpirationDateTime string      `json:"expirationDateTime"` // Required. ISO-8601 date time when the order will expire
	CompletedDateTime  string      `json:"completedDateTime"`  // Optional. ISO-8601 date time when the order was PAID/EXPIRED/ETC
	PayLink            string      `json:"payLink"`            // Required. URL to show payer on Store`s side
}

func (c *Client) CreateOrder(ctx context.Context, orderRequest OrderRequest) (*OrderPreview, error) {
	orderRequestReader, err := c.encodeRequest(orderRequest)
	var responseBodeReader io.ReadCloser
	responseBodeReader, err = c.request(ctx, c.getRequestUrl(orderPath), orderRequestReader, nil)
	if err != nil {
		return nil, err
	}
	defer func(responseBodeReader io.ReadCloser) { _ = responseBodeReader.Close() }(responseBodeReader)

	var response OrderResponse
	if err = c.decodeResponse(responseBodeReader, &response); err != nil {
		return nil, err
	}
	switch response.Status {
	case OrderStatusSuccess, OrderStatusAlready:
		return &response.Preview, nil
	default:
		return nil, fmt.Errorf("error creating order: status=%s, message=%s", response.Status, response.Message)
	}
}

func (c *Client) GetPreviewOrder(ctx context.Context, id string) (*OrderPreview, error) {
	responseBodeReader, err := c.request(ctx, c.getRequestUrl(previewPath), nil, func(q url.Values) url.Values {
		q.Add("id", id)
		return q
	})
	if err != nil {
		return nil, err
	}
	defer func(responseBodeReader io.ReadCloser) { _ = responseBodeReader.Close() }(responseBodeReader)

	var response OrderResponse
	if err = c.decodeResponse(responseBodeReader, &response); err != nil {
		return nil, err
	}
	switch response.Status {
	case OrderStatusSuccess:
		return &response.Preview, nil
	default:
		return nil, fmt.Errorf("error getting order: status=%s, message=%s", response.Status, response.Message)
	}

}
