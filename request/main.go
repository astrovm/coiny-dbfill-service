package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func dbFill(url string, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		log.Println("got error getting url:")
		log.Println(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("got error reading response body:")
		log.Println(err)
		return
	}

	type Item struct {
		URL  string
		Body string
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// create dynamoDB client
	svc := dynamodb.New(sess)

	item := Item{
		URL:  string(url),
		Body: string(body),
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		log.Println("got error marshalling new URL item:")
		log.Println(err)
		return
	}

	tableName := "coiny-apis-data"

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		log.Println("got error calling PutItem:")
		log.Println(err)
		return
	}

	log.Println("successfully added '" + item.URL + "' to table " + tableName)

	log.Println("goroutine exit")
}

func getUrls() {
	urls := []string{
		"https://blockstream.info/api/fee-estimates",
		"https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&&vs_currencies=usd",
		"https://blockstream.info/api/blocks/tip/height",
		"https://www.bitgo.com/api/v1/tx/fee",
		"https://api.smartbit.com.au/v1/blockchain/chart/block-size-total?from=2020-01-01",
		"https://ripio.com/api/v3/rates/?country=AR",
		"https://be.buenbit.com/api/market/tickers/",
		"https://api.bitso.com/v3/ticker",
		"https://argenbtc.com/public/cotizacion_js.php",
		"https://api.satoshitango.com/v3/ticker/ARS",
		"https://api.cryptomkt.com/v1/ticker",
		"https://bitex.la/api/tickers/btc_ars",
		"https://www.buda.com/api/v2/markets/btc-ars/order_book",
		"https://www.buda.com/api/v2/markets/eth-ars/order_book",
		"https://www.buda.com/api/v2/markets/ltc-ars/order_book",
		"https://www.qubit.com.ar/c_unvalue",
		"https://www.qubit.com.ar/c_value",
		"https://api.universalcoins.net/Criptomonedas/obtenerDatosHome",
		"https://api.pro.coinbase.com/products/btc-usdc/ticker",
		"https://api.pro.coinbase.com/products/dai-usdc/ticker",
		"https://api.pro.coinbase.com/products/eth-dai/ticker",
		"https://api.pro.coinbase.com/products/eth-btc/ticker",
		"https://api.pro.coinbase.com/products/ltc-btc/ticker",
		"https://api.pro.coinbase.com/products/xrp-btc/ticker",
		"https://api.pro.coinbase.com/products/xlm-btc/ticker",
	}

	urlsLength := len(urls)

	var wg sync.WaitGroup

	wg.Add(urlsLength)

	for i := 0; i < urlsLength; i++ {
		go dbFill(urls[i], &wg)
	}

	wg.Wait()

	log.Println("main exit")
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	getUrls()

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "done.",
	}, nil
}

func main() {
	lambda.Start(handler)
}
