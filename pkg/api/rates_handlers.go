package api

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/tonkeeper/opentonapi/pkg/oas"
)

func (h *Handler) GetRates(ctx context.Context, params oas.GetRatesParams) (res oas.GetRatesRes, err error) {
	params.Tokens = strings.TrimSpace(params.Tokens)
	tokens := strings.Split(params.Tokens, ",")
	if len(tokens) == 0 {
		return &oas.BadRequest{"tokens is required param"}, nil
	}

	params.Currencies = strings.TrimSpace(strings.ToUpper(params.Currencies))
	currencies := strings.Split(params.Currencies, ",")
	if len(currencies) == 0 {
		return &oas.BadRequest{"currencies is required param"}, nil
	}

	if len(params.Tokens) > 50 || len(params.Currencies) > 50 {
		return &oas.BadRequest{"max params limit is 50 items"}, nil
	}

	rates := h.tonRates.GetRates()
	pools := h.tonRates.GetPools()

	basicTonPrice := rates["TON"]

	ratesRes := make(map[string]map[string]map[string]interface{})

	for _, token := range tokens {
		for _, currency := range currencies {
			tonPriceToCurrency, ok := rates[currency]
			if !ok {
				return &oas.BadRequest{Error: "invalid currency: " + currency}, nil
			}

			price := tonPriceToCurrency
			if token != "ton" {
				tokenPrice, ok := pools[token]
				if !ok {
					return &oas.BadRequest{Error: "invalid token: " + token}, nil
				}
				if currency != "TON" {
					tokenPrice = tokenPrice / basicTonPrice
					price = tokenPrice * tonPriceToCurrency
				} else {
					price = tokenPrice / tonPriceToCurrency
				}
			}
			rate, ok := ratesRes[token]
			if !ok {
				ratesRes[token] = map[string]map[string]interface{}{"prices": {currency: price}}
				continue
			}

			rate["prices"][currency] = price
			ratesRes[token] = rate
		}
	}

	bytesRatesRes, err := json.Marshal(ratesRes)
	if err != nil {
		return &oas.InternalError{Error: err.Error()}, nil
	}

	return &oas.GetRatesOK{Rates: bytesRatesRes}, nil
}
