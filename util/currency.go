package util

// Constants for all supported currencies
const (
	GBP = "GBP"
	HKD = "HKD"
	USD = "USD"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case GBP, HKD, USD:
		return true
	default:
		return false
	}
}
