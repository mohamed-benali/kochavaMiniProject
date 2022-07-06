package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"net/http"
	"net/url"
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
)

/**
Returns the filled URL and the http method from the http body
*/
func fillURLParameters(body []byte) (method string, urlStr string) {
	var bodyMap map[string]any
	error := json.Unmarshal(body, &bodyMap)

	if error != nil {
		fmt.Println(error)
	}

	endpoint := bodyMap["endpoint"].(map[string]any)
	urlStr = endpoint["url"].(string)
	method = endpoint["method"].(string)

	data := bodyMap["data"].(map[string]any)
	for key, value := range data { // For each given query parameter replace it with the given data value
		valueToReplace := "{" + key + "}"
		nReplacements := -1 // -1 because replace all
		urlStr = strings.Replace(urlStr, valueToReplace, url.QueryEscape(value.(string)), nReplacements)
	}

	urlStr = replaceDefault(urlStr, DEFAULT_UNMATCHED_PARAMETER) // Set unmatched query parameters with default value

	return method, urlStr
}

func replaceDefault(input string, replace string) string {
	sampleRegexp := regexp.MustCompile(`{.*}`) // Regex to match all substrings between '{' and '}' included
	result := sampleRegexp.ReplaceAllString(input, replace)
	return result
}

func sendRequest(method string, url string) bool {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return false
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println(string(body))
	return true
}

func consume() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddress},
		GroupID: GroupID,
		Topic:   topic})
	i := 0
	for { // Infinite loop
		// the `ReadMessage` method blocks until we receive the next event
		msg, err := r.ReadMessage(context.Background())
		if err != nil {
			panic("could not read message " + err.Error())
		}
		// after receiving the message, log its value

		method, urlString := fillURLParameters(msg.Value)

		successDeliver := false
		deliverTries := 0
		for !successDeliver && deliverTries < RETRY_ATTEMPTS { // Retry if not delivered and less than RETRY_ATTEMPTS tries
			deliverTries++
			successDeliver = sendRequest(method, urlString)
			time.Sleep(RETRY_WAIT_TIME * time.Second)
		}

		fmt.Println("received: ", string(msg.Value))
		fmt.Println(i, "\n")
		i++
	}
}

func main() {
	fmt.Println("Hello, 世界")
	consume()
}
