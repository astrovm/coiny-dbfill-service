package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

func getURLData(url string) (map[string]interface{}, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	readResponse, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var body map[string]interface{}

	json.Unmarshal([]byte(readResponse), &body)

	return body, nil
}

func getAndSave(url string, cache map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// some external apis we call have rate limits, so we sleep for 1 sec to respect them
	if url == "https://api.pro.coinbase.com/products/xrp-btc/ticker" {
		time.Sleep(1 * time.Second)
	}

	body, err := getURLData(url)
	if err != nil {
		log.Println(body)
		log.Println(err)
		return
	}

	cache[url] = body
}

func triggerURLsGets(cache *map[string]interface{}) {
	urlsArray := []string{
		"https://api.bitso.com/v3/ticker",
		"https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&&vs_currencies=usd",
		"https://api.cryptomkt.com/v1/ticker",
		"https://api.pro.coinbase.com/products/btc-usdc/ticker",
		"https://api.pro.coinbase.com/products/dai-usdc/ticker",
		"https://api.pro.coinbase.com/products/eth-btc/ticker",
		"https://api.pro.coinbase.com/products/eth-dai/ticker",
		"https://api.pro.coinbase.com/products/ltc-btc/ticker",
		"https://api.pro.coinbase.com/products/xlm-btc/ticker",
		"https://api.pro.coinbase.com/products/xrp-btc/ticker",
		"https://api.satoshitango.com/v3/ticker/ARS",
		"https://api.smartbit.com.au/v1/blockchain/chart/block-size-total?from=2020-01-01",
		"https://api.universalcoins.net/Criptomonedas/obtenerDatosHome",
		"https://argenbtc.com/public/cotizacion_js.php",
		"https://be.buenbit.com/api/market/tickers/",
		"https://bitex.la/api/tickers/btc_ars",
		"https://blockstream.info/api/blocks/tip/height",
		"https://blockstream.info/api/fee-estimates",
		"https://ripio.com/api/v3/rates/?country=AR",
		"https://www.bitgo.com/api/v1/tx/fee",
		"https://www.buda.com/api/v2/markets/btc-ars/order_book",
		"https://www.buda.com/api/v2/markets/eth-ars/order_book",
		"https://www.buda.com/api/v2/markets/ltc-ars/order_book",
		"https://www.qubit.com.ar/c_unvalue",
		"https://www.qubit.com.ar/c_value",
	}

	urlsLength := len(urlsArray)

	var wg sync.WaitGroup

	wg.Add(urlsLength)

	for i := 0; i < urlsLength; i++ {
		go getAndSave(urlsArray[i], *cache, &wg)
	}

	wg.Wait()
}

func main() {
	cache := make(map[string]interface{})

	go triggerURLsGets(&cache)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString("asd"))
		log.Println(cache)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))

}
