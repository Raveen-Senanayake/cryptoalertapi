package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CoinGeckoTicker struct {
	Base string `json:"base"` 
	Target string `json:"target"`  
	Last float64 `json:"last"`
	Volume float64 `json:"volume"`
	CoinId string `json:"coin_id"`
	TargetCoinId string `json:"target_coin_id"`
}

type CoinGeckoReturnObject struct {
	Name 	string 	`json:"name"`
	Ticker	[]CoinGeckoTicker `json:"tickers"`

}

//check if coingecko currency has required object , if so return the object if not return in USD
func analyseCoinGeckoReturn(coin_gecko_return_object CoinGeckoReturnObject , fiat string) (CoinGeckoTicker , string) {

	var converted_cryto_object CoinGeckoTicker 
	var converted_crypto_object_usdt CoinGeckoTicker
	return_fiat_type := "USDT"
	ticker_array := coin_gecko_return_object.Ticker	
	

	for _ , s:= range ticker_array {
		if s.Target == fiat {
			converted_cryto_object = s
			return converted_cryto_object , s.Base
		}
		if s.Target == "USDT" {
			converted_crypto_object_usdt  = s	
		}
	}
	if converted_cryto_object == (CoinGeckoTicker{}) {
		converted_cryto_object = converted_crypto_object_usdt 
	}

	return converted_cryto_object, return_fiat_type
}


// convert currency to the required balye

func convertToRequiredFiatCurrency(usdt_value float64, required_fiat_type string )  int  {

	currency_exchange_api_call_link := "https://currency-exchange.p.rapidapi.com/exchange?to="+required_fiat_type+"&from=USD&Q=";


}


func getCurrencyCurrentPrice(c* gin.Context) {
	id := c.Param("id")
	fiat := c.Query("fiat")
	exchange := c.Query("exchange")

	 if len(id) == 0 || len(fiat)== 0 || len(exchange) ==0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"paramters" : "parameters missing"})
	 }
	 // Url coin gecko url
	 coin_gecko_call_url :="https://api.coingecko.com/api/v3/exchanges/"+exchange+"/tickers?coin_ids="+id

	//HTTP CALL
	 resp, err := http.Get(coin_gecko_call_url)
	if err != nil {
		fmt.Println("DJDJD")
		c.IndentedJSON(http.StatusNotFound, gin.H{"coingecko" : "error"})
	}

	defer resp.Body.Close()

	//Decode JSON 
	coin_gecko_return_object := &CoinGeckoReturnObject{}
	dec := json.NewDecoder(resp.Body)
	
	if err:= dec.Decode(coin_gecko_return_object); err !=nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"coingecko" : "Decode error"})
	} else {
	
		return_ticker_object := *coin_gecko_return_object
		crypto_object_to_return , extracted_fiat := analyseCoinGeckoReturn(return_ticker_object, fiat)

		if extracted_fiat != fiat {
			if extracted_fiat == "USDT" {
				currency_usd_value := crypto_object_to_return.Last
				required_fiat_type := fiat
				converted_to_required_value :=  convertToRequiredFiatCurrency(currency_usd_value , required_fiat_type ) 
				fmt.Println(converted_to_required_value)		
			}
		} 

		fmt.Println(crypto_object_to_return , extracted_fiat)
		c.IndentedJSON(http.StatusOK, coin_gecko_return_object)

	}
}


func main() {

	router := gin.Default()
	router.GET("/currency/:id" , getCurrencyCurrentPrice)
    router.Run("localhost:8080")

}
