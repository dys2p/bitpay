package bitpay

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/bitpay/bitpay-go/key_utils"
)

type Client struct {
	API string // URI
	ID  string // SIN
	Key string
}

func MakeClient(api string, keyPath string) (*Client, error) {

	// check if key file exists

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		var key = key_utils.GeneratePem()
		if err := ioutil.WriteFile(keyPath, []byte(key), 0400); err != nil {
			return nil, fmt.Errorf("creating keyfile %s: %w", keyPath, err)
		}
		log.Printf("created keyfile: %s", keyPath)
	}

	// read key file

	b, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading keyfile %s: %w", keyPath, err)
	}
	key := string(b)

	// print SIN

	log.Printf("please make sure that your BTCPay store is paired to the public key (hex SIN): %s", key_utils.ExtractCompressedPublicKey(key))

	return &Client{
		API: api,
		ID:  key_utils.GenerateSinFromPem(key),
		Key: key,
	}, nil
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
	return path.Join(client.API, "i", invoice.ID) // don't use invoice.URL because that's for logged-in btcpay users only
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
	CryptoCode string // like "BTC" or "XMR"
	Paid       string // float64 won't work, server returns a string
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
