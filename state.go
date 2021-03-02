package bitpay

// See https://bitpay.com/docs/invoice-states
type InvoiceState string

const (
	New         InvoiceState = "new"
	Paid                     = "paid" // "amount paid is equal to or greater than the amount expected", but not confirmed yet
	Confirmed                = "confirmed"
	Complete                 = "complete"
	Expired                  = "expired"
	Invalid                  = "invalid"
	PaidPartial              = "paidPartial"
	PaidOver                 = "paidOver"
)
