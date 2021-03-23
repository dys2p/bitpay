package bitpay

// See https://bitpay.com/docs/invoice-states
type InvoiceState string

const (
	New         InvoiceState = "new"
	Paid        InvoiceState = "paid" // "amount paid is equal to or greater than the amount expected", but not confirmed yet
	Confirmed   InvoiceState = "confirmed"
	Complete    InvoiceState = "complete"
	Expired     InvoiceState = "expired"
	Invalid     InvoiceState = "invalid"
	PaidPartial InvoiceState = "paidPartial"
	PaidOver    InvoiceState = "paidOver"
)
