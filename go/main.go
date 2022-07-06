package main

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	topic                       = "postback"
	brokerAddress               = "localhost:9092"
	GroupID                     = "Group"
	DEFAULT_UNMATCHED_PARAMETER = "UNMATCHED"
	RETRY_ATTEMPTS              = 3
	RETRY_WAIT_TIME             = 1 // In seconds
	LOGS_ACTIVED                = true
	LOGS_FILE_NAME              = "logs_file"
)

func setLogger() {
	if LOGS_ACTIVED { // If logs are active, it will write them on the file
		f, err := os.OpenFile(LOGS_FILE_NAME, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		log.SetOutput(f)
	} else {
		log.SetOutput(ioutil.Discard)
	}
}

/**
Returns the filled URL and the http method using the http body
*/
func fillURLParameters(body []byte) (method string, urlStr string) {
	var bodyMap map[string]any
	error := json.Unmarshal(body, &bodyMap)

	if error != nil {
		log.Println(error)
	}

	endpoint := bodyMap["endpoint"].(map[string]any)
	urlStr = endpoint["url"].(string)
	method = endpoint["method"].(string)

	data := bodyMap["data"].(map[string]any) // I assume there is not an empty postback
	for key, value := range data {           // For each given query parameter replace it with the given data value
		valueToReplace := "{" + key + "}"
		nReplacements := -1 // -1 because replace all
		urlStr = strings.Replace(urlStr, valueToReplace, url.QueryEscape(value.(string)), nReplacements)
	}

	urlStr = replaceDefault(urlStr, DEFAULT_UNMATCHED_PARAMETER) // Set unmatched query parameters with default value

	return method, urlStr
}

// Replaces the non matched query parameters with replace string
func replaceDefault(input string, replace string) string {
	sampleRegexp := regexp.MustCompile(`{.*}`) // Regex to match all substrings between '{' and '}' included
	result := sampleRegexp.ReplaceAllString(input, replace)
	return result
}

// Make the HTTP request and log response data
func sendRequest(method string, url string) (success bool, res *http.Response) {
	log.Println("Send request: ", method, " ", url)
	start := time.Now()

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil) // Create request
	if err != nil {
		log.Println(err)
		return false, nil
	}

	res, err = client.Do(req) // Send request
	if err != nil {
		log.Println(err)
		return false, nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body) // Get response body
	if err != nil {
		log.Println(err)
		return false, nil
	}

	deliveryTime := time.Since(start)

	log.Println("Response log - Code: ", res.StatusCode, " - Delivery time: ", deliveryTime, " - Body: ", string(body))
	return true, res
}

func consume() {
	r := kafka.NewReader(kafka.ReaderConfig{ // Create Apache Kafka Consumer
		Brokers: []string{brokerAddress},
		GroupID: GroupID,
		Topic:   topic})
	i := 0
	for { // Infinite loop
		msg, err := r.ReadMessage(context.Background()) // the `ReadMessage` method blocks until we receive the next event
		if err != nil {
			log.Println("could not read message " + err.Error())
		}

		method, urlString := fillURLParameters(msg.Value)

		successDeliver, _ := sendRequest(method, urlString)
		deliverTries := 1
		for !successDeliver && deliverTries < RETRY_ATTEMPTS { // Retry if not delivered and less than RETRY_ATTEMPTS tries
			successDeliver, _ = sendRequest(method, urlString)
			time.Sleep(RETRY_WAIT_TIME * time.Second)
			deliverTries++
		}
		i++
	}
}

func main() {
	setLogger()
	consume()
}
