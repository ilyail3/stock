package stock

import (
	"time"
	"regexp"
	"fmt"
	"strconv"
	"net/http"
	"encoding/json"
	"sort"
)

type DataPoint struct {
	Date time.Time

	Low float64
	High float64

	Open float64
	Close float64

	Volume float64
}

type Reader interface {
	GetPrice(symbol string) ([]DataPoint, error)
}

type aVantageMetadata struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	Interval      string `json:"4. Interval"`
	OutputSize    string `json:"5. Output Size"`
	TimeZone      string `json:"6. Time Zone"`
}

type aVantageResponse struct {
	Metadata aVantageMetadata `json:"Meta Data"`
	TimeSeries map[string]map[string]string `json:"Time Series (15min)"`
}

type alphaVantageReader struct {
	apiKey string
}

func priceParser(err error, columnName string) (func(error,map[string]string)(float64,error), error) {
	if err != nil {
		return nil, err
	}

	r,err := regexp.Compile("^[0-9]+\\. " + regexp.QuoteMeta(columnName) + "$")
	if err != nil {
		return nil, fmt.Errorf("failed to compile price parser")
	}

	return func (err error, dataMap map[string]string) (float64, error){
		if err != nil {
			return 0, err
		}

		for key, obj := range dataMap{
			if r.MatchString(key) {
				v, err := strconv.ParseFloat(obj, 64)

				if err != nil {
					return 0, fmt.Errorf("failed to parse %s, value %s", columnName, obj)
				}

				return v, nil
			}
		}

		return 0, fmt.Errorf("missing column for %s", columnName)
	}, nil
}

func reverseStrings(input []string) []string {
	if len(input) == 0 {
		return input
	}
	return append(reverseStrings(input[1:]), input[0])
}

func (reader *alphaVantageReader) GetPrice(symbol string) ([]DataPoint, error){
	url := fmt.Sprintf(
		"https://www.alphavantage.co/query?function=TIME_SERIES_INTRADAY&symbol=%s&interval=15min&apikey=%s",
		symbol,
		reader.apiKey)

	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to query alphavantage")
	}

	decoder := json.NewDecoder(resp.Body)
	respObj := &aVantageResponse{}

	err = decoder.Decode(respObj)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response from alphavantage")
	}

	tz,err := time.LoadLocation(respObj.Metadata.TimeZone)

	if err != nil {
		return nil, fmt.Errorf("cannot load timezone for:%s", respObj.Metadata.TimeZone)
	}

	result := make([]DataPoint, len(respObj.TimeSeries))

	closeParser,err := priceParser(err, "close")
	openParser,err := priceParser(err, "open")

	highParser,err := priceParser(err, "high")
	lowParser,err := priceParser(err, "low")

	volumeParser,err := priceParser(err, "volume")

	if err != nil {
		return nil, err
	}

	keys := make([]string, len(respObj.TimeSeries))

	i := 0
	for key := range respObj.TimeSeries{
		keys[i] = key
		i ++
	}

	sort.Strings(keys)

	for i, key := range reverseStrings(keys) {
		t,err := time.ParseInLocation("2006-01-02 15:04:05", key, tz)
		obj := respObj.TimeSeries[key]

		if err != nil {
			return nil, fmt.Errorf("failed to parse date:%s", key)
		}

		result[i].Date = t.UTC()

		result[i].Close,err = closeParser(err, obj)
		result[i].Open,err = openParser(err, obj)

		result[i].High,err = highParser(err, obj)
		result[i].Low,err = lowParser(err, obj)

		result[i].Volume,err = volumeParser(err, obj)

		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func AlphaAdvantageReader(apiKey string) Reader{
	return &alphaVantageReader{ apiKey:apiKey }
}