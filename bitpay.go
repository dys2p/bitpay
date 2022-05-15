package bitpay

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/bitpay/bitpay-go/key_utils"
)

type Client struct {
	API string `json:"api"` // URI
	Key string `json:"key"`
}

// LoadClient reads an API key from a file. If the file doesn't exist, a key is generated and written to it.
func LoadClient(jsonPath string) (*Client, error) {
	var client = &Client{}
	data, err := os.ReadFile(jsonPath)
	switch {
	case err == nil:
		return client, json.Unmarshal(data, client)
	case os.IsNotExist(err):
		return nil, CreateClientConfig(jsonPath)
	default:
		return nil, err
	}
}

// CreateClientConfig creates an empty json config file with empty values and chmod 600, so someone can fill in easily.
// CreateClientConfig always returns an error.
func CreateClientConfig(jsonPath string) error {
	data, err := json.Marshal(&Client{
		Key: key_utils.GeneratePem(),
	})
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, data, 0600); err != nil {
		return err
	}
	return fmt.Errorf("created empty config file: %s", jsonPath)
}

func (client *Client) DoRequest(method string, path string, body io.Reader) (*http.Response, error) {

	uri := client.API + "/" + path

	req, err := http.NewRequest(http.MethodGet, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Identity", key_utils.ExtractCompressedPublicKey(client.Key))
	req.Header.Set("X-Signature", key_utils.Sign(uri, client.Key))

	return (&http.Client{
		Timeout: 10 * time.Second,
	}).Do(req)
}

func (client *Client) GetInvoice(invID string) (*Invoice, error) {

	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("invoices/%s?token=%s", invID, "irrelevant"), nil) // merchant token could be obtained with GetTokens
	if err != nil {
		return nil, fmt.Errorf("getting invoice: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	iw := &invoiceWrapper{}
	if err := json.Unmarshal(body, iw); err == nil {
		return &iw.Data, nil
	} else {
		return nil, err
	}
}

// InvoiceURL returns an absolute URL.
func (client *Client) InvoiceURL(invoice *Invoice) string {
	return client.API + "/i/" + invoice.ID // don't use invoice.URL because that's for logged-in btcpay users only
}

// SINHex returns the hex representation of the public key.
// Use this for pairing with BTCPay Server.
func (client *Client) SINHex() string {
	return key_utils.ExtractCompressedPublicKey(client.Key)
}

// required for json unmarshal because invoice is wrapped in a data field
type invoiceWrapper struct {
	Data Invoice
}

type Invoice struct {
	CryptoInfo     []CryptoInfo
	Currency       string
	ExpirationTime int64 // milliseconds
	ID             string
	OrderID        string
	Price          float64
	Status         string
}

func (invoice *Invoice) Expiration() time.Time {
	return time.Unix(invoice.ExpirationTime/1000, 0)
}

type CryptoInfo struct {
	CryptoCode string  // like "BTC" or "XMR"
	Paid       float64 `json:",string"`
	Payments   []Payment
	Rate       float64 // from the time of invoice generation
}

type Payment struct {
	Completed    bool // based on the current time, not on the invoice expiry
	Confirmed    bool // based on the current time, not on the invoice expiry
	Fee          float64
	ID           string // store it in your database to avoid double booking if the webhook is called twice (e.g. if an invoice receives multiple payments)
	ReceivedDate string
	Value        float64
}

func (payment *Payment) ParseReceivedDate() (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05.999", payment.ReceivedDate)
	if err != nil {
		err = fmt.Errorf("error parsing payment receive date %s: %v", payment.ReceivedDate, err)
	}
	return t, err
}
