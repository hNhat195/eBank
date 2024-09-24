package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/nhat195/simple_bank/util"
)

var validatorCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}
	return false
}
