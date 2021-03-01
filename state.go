package bitpay

// See https://bitpay.com/docs/invoice-states
type InvoiceState string

const (
	New         InvoiceState = "new"
	Paid        string       = "paid" // "amount paid is equal to or greater than the amount expected", but not confirmed yet
	Confirmed   string       = "confirmed"
	Complete    string       = "complete"
	Expired     string       = "expired"
	Invalid     string       = "invalid"
	PaidPartial string       = "paidPartial"
	PaidOver    string       = "paidOver"
)
