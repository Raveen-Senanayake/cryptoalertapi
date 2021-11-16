package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/apex/gateway"
	"github.com/gin-gonic/gin"
)

func inLambda() bool {
	if lambdaTaskRoot := os.Getenv("LAMBDA_TASK_ROOT"); lambdaTaskRoot != "" {
		return true
	}
	return false
}

//CoinGeckoTicker is a struct that provides a structure to repackage infomation returned from the CoinGecko Api
type CoinGeckoTicker struct {
	Base         string  `json:"base"`
	Target       string  `json:"target"`
	Last         float64 `json:"last"`
	Volume       float64 `json:"volume"`
	CoinID       string  `json:"coin_id"`
	TargetCoinID string  `json:"target_coin_id"`
}

//CoinGeckoReturnObject is a struct that is used implement the return object from CoinGecko Api
type CoinGeckoReturnObject struct {
	Name   string            `json:"name"`
	Ticker []CoinGeckoTicker `json:"tickers"`
}

// Simple Api calls
type SimpleCoinGeckoCall struct {
	CustomFields map[string]interface{}
}

// UpdateLast update the currenacy value of the object
func UpdateLast(updateItem *CoinGeckoTicker, currencyValue float64) {
	updateItem.Last = currencyValue
}

//UpdateTarget upates the target value of a coingecko object
func UpdateTarget(updateItem *CoinGeckoTicker, currencyName string) {
	updateItem.Target = currencyName

}

//check if coingecko currency has required object , if so return the object if not return in USD
func analyseCoinGeckoReturn(coinGeckoReturnObject CoinGeckoReturnObject, fiat string) (CoinGeckoTicker, string) {

	var convertedCrytoObject CoinGeckoTicker
	var convertedCryptoObjectUsdt CoinGeckoTicker
	returnFiatType := "USDT"
	tickerArray := coinGeckoReturnObject.Ticker

	for _, s := range tickerArray {
		if s.Target == fiat {
			convertedCrytoObject = s
			return convertedCrytoObject, s.Target
		}
		if s.Target == "USD" {
			convertedCryptoObjectUsdt = s
		} else if s.Target == "USDT" {
			convertedCryptoObjectUsdt = s
		} else if s.Target == "USDC" {
			convertedCryptoObjectUsdt = s
		}
	}
	if convertedCrytoObject == (CoinGeckoTicker{}) {
		convertedCrytoObject = convertedCryptoObjectUsdt
	}
	return convertedCrytoObject, returnFiatType
}

// convert fiat currency
func convertToRequiredFiatCurrency(usdtValue float64, requiredFiatType string, c *gin.Context) float64 {

	currencyExchangeAPICALLLink := "https://currency-exchange.p.rapidapi.com/exchange?to=" + requiredFiatType + "&from=USD=1"

	// make api call
	req, _ := http.NewRequest("GET", currencyExchangeAPICALLLink, nil)

	req.Header.Add("x-rapidapi-host", "currency-exchange.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", "e152fe0ac7msh82be55889f4e392p160cd0jsn2ee8bf76ee34")
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"currency change": "currency change"})
	}

	defer res.Body.Close()

	rate, _ := ioutil.ReadAll(res.Body)

	rateInString := string(rate)
	rateInNumber, _ := strconv.ParseFloat(rateInString, 64)

	convertedValue := rateInNumber * usdtValue

	return convertedValue
}

// Gets a single coins price in desired currency
func getCoinGeckoUnitPrice(id string, exchange string, fiat string, c *gin.Context) (coinInfo *CoinGeckoTicker, err error) {

	coinGeckoCallURL := "https://api.coingecko.com/api/v3/exchanges/" + exchange + "/tickers?coin_ids=" + id

	response, err := http.Get(coinGeckoCallURL)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"coingecko": "Coin Gecko Url Error"})

		return nil, err

	}

	defer response.Body.Close()

	coinGeckoReturnObject := &CoinGeckoReturnObject{}
	dec := json.NewDecoder(response.Body)

	if err := dec.Decode(coinGeckoReturnObject); err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"coingecko": "Decode error"})
		return nil, err
	}

	returnTickerObject := *coinGeckoReturnObject

	if len(returnTickerObject.Ticker) == 0 {
		baseCoinGeckoUrl := "https://api.coingecko.com/api/v3/simple/price?ids=" + id + "&vs_currencies=" + fiat
		response, err := http.Get(baseCoinGeckoUrl)

		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"coingecko": "Coin Gecko Url Error", "Error": err.Error()})
			return nil, err

		}
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"coingecko": "general coin price indicator error"})
			return nil, err
		}

		jsonMap := make(map[string]interface{})
		erro := json.Unmarshal([]byte(bodyBytes), &jsonMap)
		if erro != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"UnpackJsonMap": "general coin price indicator error"})
			return nil, err
		}

		// extract price
		var price float64
		for _, value := range jsonMap {
			hh := value.(interface{}).(map[string]interface{})
			price = hh[strings.ToLower(fiat)].(float64)

		}

		newCoinGeckoTicker := CoinGeckoTicker{id, strings.ToUpper(fiat), price, 0, id, ""}
		arrayToInsert := make([]CoinGeckoTicker, 1)
		arrayToInsert[0] = newCoinGeckoTicker
		returnTickerObject = CoinGeckoReturnObject{"FTX", arrayToInsert}

	}

	cryptoObjectToReturn, extractedFiat := analyseCoinGeckoReturn(returnTickerObject, fiat)

	if extractedFiat != fiat {
		if extractedFiat == "USDT" || extractedFiat == "USD" || extractedFiat == "USDC" {
			currencyUsdValue := cryptoObjectToReturn.Last
			requiredFiatType := fiat

			convertedToRequiredValue := convertToRequiredFiatCurrency(currencyUsdValue, requiredFiatType, c)

			// update object
			UpdateLast(&cryptoObjectToReturn, convertedToRequiredValue)
			UpdateTarget(&cryptoObjectToReturn, fiat)

		}
	}

	return &cryptoObjectToReturn, nil
}

// Url routed view function
func getCurrencyCurrentPrice(c *gin.Context) {

	fiat := c.Query("fiat")

	exchangeList := strings.Split(c.Query("exchangelist"), ",")
	cryptoList := strings.Split(c.Query("cryptolist"), ",")

	if len(exchangeList) == 0 || len(fiat) == 0 || len(cryptoList) == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"paramters": "parameters missing"})
	}

	// send multiple request

	coinMap := make(map[int]*CoinGeckoTicker, len(cryptoList))

	wg := sync.WaitGroup{}

	for i := 0; i < len(cryptoList); i++ {

		wg.Add(1)

		go func(i int) {
			id := cryptoList[i]
			exchange := exchangeList[i]
			coin, err := getCoinGeckoUnitPrice(id, exchange, fiat, c)
			if err != nil {
				c.IndentedJSON(http.StatusNotFound, gin.H{"coingecko": "Go routine Error", "Error": err.Error()})
				return
			}
			coinMap[i] = coin

			wg.Done()

		}(i)

	}

	wg.Wait()

	c.IndentedJSON(http.StatusOK, coinMap)
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/currency", getCurrencyCurrentPrice)
	return router
}

func main() {

	if inLambda() {
		fmt.Println("running aws lambda in aws")
		log.Fatal(gateway.ListenAndServe(":8080", setupRouter()))
	} else {
		fmt.Println("running aws lambda in local")
		log.Fatal(http.ListenAndServe(":8080", setupRouter()))
	}
}
