package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/jtaylor-io/safe-as-houses/util"
	"github.com/shopspring/decimal"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}
	return false
}

var positiveDecimal validator.Func = func(fl validator.FieldLevel) bool {
	if x, ok := fl.Field().Interface().(decimal.Decimal); ok {
		return x.IsPositive()
	}
	return false
}
