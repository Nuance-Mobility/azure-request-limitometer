package main

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/Azure/go-autorest/autorest"
)

// Example Request Headers:
// 'x-ms-ratelimit-remaining-resource': 'Microsoft.Compute/HighCostGet3Min;133,Microsoft.Compute/HighCostGet30Min;657'
// 'x-ms-ratelimit-remaining-resource': 'Microsoft.Compute/LowCostGet3Min;3989,Microsoft.Compute/LowCostGet30Min;31790'
// 'x-ms-ratelimit-remaining-resource': 'Microsoft.Compute/PutVM3Min;740,Microsoft.Compute/PutVM30Min;3695'

var expectedHeaderField = "X-Ms-Ratelimit-Remaining-Resource"
var expectedHeaderFormat = regexp.MustCompile(`(Microsoft.\w+\/\w+);(\d+)`)

func getRequestsRemaining(nodename string) (requestsRemaining map[string]int) {
	requestsRemaining = make(map[string]int)

	responses := []autorest.Response{
		azureClient.GetVM(nodename).Response,
		azureClient.GetAllVM().Response().Response,
		azureClient.PutVM(nodename),
		azureClient.GetNic(nodename).Response,
		azureClient.GetAllNics().Response().Response,
		azureClient.PutNic(nodename),
	}

	for _, response := range responses {
		if response.StatusCode != 200 {
			log.Fatalf("Response did not return a StatusCode of 200. StatusCode: %d", response.StatusCode)
		}
		for k, v := range extractRequestsRemaining(response.Header) {
			requestsRemaining[k] = v
		}
	}

	return
}

func extractRequestsRemaining(h http.Header) (requestsRemaining map[string]int) {
	requestsRemaining = map[string]int{}

	headerSubfields := strings.Split(h.Get(expectedHeaderField), ",")

	for _, field := range headerSubfields {
		matches := expectedHeaderFormat.FindStringSubmatch(field)
		if !(len(matches) == 3) {
			log.Fatalf("header didn't contain expected data: %s", field)
		}

		requestType := matches[1]
		requestsLeft, err := strconv.Atoi(matches[2])
		if err != nil {
			log.Fatal(err)
		}
		requestsRemaining[requestType] = requestsLeft
	}

	return requestsRemaining
}
