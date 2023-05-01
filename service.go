package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type coinmarketcapResponse struct {
	Data map[string]struct {
		Quote map[string]struct {
			Price float64 `json:"price"`
		} `json:"quote"`
	} `json:"data"`
}

// Interface that can fetch a price
type PriceService interface {
	FetchPrice(context.Context, string) (float64, error)
}

// implements the PriceFetcher interface
type priceService struct {
}

func (s *priceService) FetchPrice(ctx context.Context, ticker string) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	q := req.URL.Query()
	q.Set("symbol", ticker)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("X-CMC_PRO_API_KEY", os.Getenv("CMC_API_KEY"))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make HTTP request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var cmcResp coinmarketcapResponse
	err = json.NewDecoder(resp.Body).Decode(&cmcResp)
	if err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	quote, ok := cmcResp.Data[ticker]
	if !ok {
		return 0, fmt.Errorf("the given ticker (%s) is not supported", ticker)
	}

	price, ok := quote.Quote["USD"]
	if !ok {
		return 0, fmt.Errorf("USD price not available for ticker: %s", ticker)
	}

	return price.Price, nil
}

type loggingService struct {
	priceService
}

func (s loggingService) FetchPrice(ctx context.Context, ticker string) (price float64, err error) {
	defer func(begin time.Time) {
		logrus.WithFields(logrus.Fields{
			"requestID": ctx.Value("requestID"),
			"took":      time.Since(begin),
			"err":       err,
			"price":     price,
		}).Info("fetchPrice")
	}(time.Now())

	return s.priceService.FetchPrice(ctx, ticker)
}
