package main

import (
	"net/http"
	"fmt"
	"encoding/json"
	"time"
	"strconv"
	"regexp"
	"os"
)

type AVantageMetadata struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	Interval      string `json:"4. Interval"`
	OutputSize    string `json:"5. Output Size"`
	TimeZone      string `json:"6. Time Zone"`
}

type AVantageResponse struct {
	Metadata AVantageMetadata `json:"Meta Data"`
	TimeSeries map[string]map[string]string `json:"Time Series (15min)"`
}

type ApplicationConfig interface {
	GetApiKey() string
}

type JsonConfig struct {
	ApiKey string
}

func(j *JsonConfig) GetApiKey() string{
	return j.ApiKey
}

type DataPoint struct {
	Date time.Time

	Low float64
	High float64

	Open float64
	Close float64

	Volume float64
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
				v, lerr := strconv.ParseFloat(obj, 64)

				if lerr != nil {
					return 0, fmt.Errorf("Failed to parse %s, value %s", columnName, obj, lerr)

				}

				return v, nil
			}
		}

		return 0, fmt.Errorf("missing column for %s", columnName, nil)
	}, nil
}

func GetPrice(config ApplicationConfig, symbol string) ([]DataPoint, error) {
	url := fmt.Sprintf(
		"https://www.alphavantage.co/query?function=TIME_SERIES_INTRADAY&symbol=%s&interval=15min&apikey=%s",
		symbol,
		config.GetApiKey())

	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to query alphavantage", err)
	}

	decoder := json.NewDecoder(resp.Body)
	resp_obj := &AVantageResponse{}

	err = decoder.Decode(resp_obj)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response from alphavantage", err)
	}

	tz,err := time.LoadLocation(resp_obj.Metadata.TimeZone)

	if err != nil {
		return nil, fmt.Errorf("cannot load timezone for:%s", resp_obj.Metadata.TimeZone, err)
	}

	result := make([]DataPoint, len(resp_obj.TimeSeries))
	i := 0

	closeParser,err := priceParser(err, "close")
	openParser,err := priceParser(err, "open")

	highParser,err := priceParser(err, "high")
	lowParser,err := priceParser(err, "low")

	volumeParser,err := priceParser(err, "volume")

	if err != nil {
		return nil, err
	}

	for key, obj := range resp_obj.TimeSeries {
		t,err := time.ParseInLocation("2006-01-02 15:04:05", key, tz)

		if err != nil {
			return nil, fmt.Errorf("failed to parse date:%s", key, err)
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

		i++
	}

	return result, nil
}

func readConfig(fileName string) (ApplicationConfig,error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	conf := &JsonConfig{}
	decoder := json.NewDecoder(f)
	err = decoder.Decode(conf)

	if err != nil {
		return nil, err
	}
	return conf, nil
}


func main(){
	config, err := readConfig("config.json")
	fmt.Printf("api key:%s", config.GetApiKey())

	if err != nil {
		panic(err)
	}

	r, err := GetPrice(config, "TSLA")

	if err != nil {
		panic(err)
	}

	fmt.Println(r)
}