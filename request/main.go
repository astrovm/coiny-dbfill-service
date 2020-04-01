package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-lambda-go/lambda"
)

func getURLData(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "got error getting url:", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "got error reading response body:", err
	}

	return string(body), nil
}

func dbFillWithURLData(url string, body string) (string, error) {
	type Item struct {
		URL  string
		Body string
	}

	itemToAddtoDB := Item{
		URL:  string(url),
		Body: string(body),
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(itemToAddtoDB)
	if err != nil {
		return "got error marshalling new URL item:", err
	}

	tableName := "coiny-apis-data"

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return "got error calling PutItem:", err
	}

	return "successfully added '" + itemToAddtoDB.URL + "' to table " + tableName, nil
}

func getAndSave(url string, wg *sync.WaitGroup) {
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

	res, err := dbFillWithURLData(url, body)
	if err != nil {
		log.Println(res)
		log.Println(err)
		return
	}

	log.Println(res)
}

func triggerURLsGets() {
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
		go getAndSave(urlsArray[i], &wg)
	}

	wg.Wait()
}

func handler() (string, error) {
	triggerURLsGets()

	return "done", nil
}

func main() {
	lambda.Start(handler)
}
