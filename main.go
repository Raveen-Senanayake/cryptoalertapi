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

func getCurrencyCurrentPrice(c* gin.Context) {
	id := c.Param("id")
	fiat := c.Query("fiat")
	exchange := c.Query("exchange")

	 if len(id) == 0 || len(fiat)== 0 || len(exchange) ==0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"paramters" : "parameters missing"})
	 }
	 // Url coin gecko url
	 coin_gecko_call_url :="https://api.coingecko.com/api/v3/exchanges/"+exchange+"/tickers?coin_ids="+id

	 fmt.Println(coin_gecko_call_url)

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
		fmt.Println(err)
		c.IndentedJSON(http.StatusNotFound, gin.H{"coingecko" : "Decode error"})
	} else {
		c.IndentedJSON(http.StatusOK, coin_gecko_return_object)

	}
}


func main() {

	router := gin.Default()
	router.GET("/currency/:id" , getCurrencyCurrentPrice)

    router.Run("localhost:8080")


}
