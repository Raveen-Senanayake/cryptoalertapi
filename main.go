package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

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

func UpdateLast (updateItem CoinGeckoTicker , currencyValue float64) CoinGeckoTicker {
	updateItem.Last = currencyValue
	return updateItem
	
}

func UpdateTarget  (updateItem CoinGeckoTicker , currencyName string) CoinGeckoTicker {
	updateItem.Target = currencyName
	return updateItem
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
			return converted_cryto_object , s.Target
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


// convert fiat currency
func convertToRequiredFiatCurrency(usdt_value float64, required_fiat_type string )  float64  {
	
	s := fmt.Sprintf("%f", usdt_value)
	currency_exchange_api_call_link := "https://currency-exchange.p.rapidapi.com/exchange?to="+required_fiat_type+"&from=USD&Q=1"+s;
	
	// make api call
	req, _ := http.NewRequest("GET", currency_exchange_api_call_link , nil)

	req.Header.Add("x-rapidapi-host", "currency-exchange.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", "e152fe0ac7msh82be55889f4e392p160cd0jsn2ee8bf76ee34")
	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	rate, _ := ioutil.ReadAll(res.Body)

	rate_in_string := string(rate)
	rate_in_number, _ := strconv.ParseFloat(rate_in_string,64)

	converted_value := rate_in_number * usdt_value

	return converted_value
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
				

				// update object
				itemWithLastUpdated := UpdateLast(crypto_object_to_return, converted_to_required_value )
				itemWithTargetUpdated := UpdateTarget(itemWithLastUpdated , fiat)
				c.IndentedJSON(http.StatusOK, itemWithTargetUpdated)

			} ;
		} else {
			c.IndentedJSON(http.StatusOK, crypto_object_to_return)

		}
	}
}


func main() {

	router := gin.Default()
	router.GET("/currency/:id" , getCurrencyCurrentPrice)

	port := os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}
    router.Run(":"+port)

}
