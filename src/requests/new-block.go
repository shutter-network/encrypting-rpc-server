package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type RequestPayload struct {
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
	JsonRPC string        `json:"jsonrpc"`
}

func SendNewBlock(blockNumber uint64, httpListenAddress string) {
	payload := RequestPayload{
		Method:  "eth_newBlock",
		Params:  []interface{}{blockNumber},
		ID:      1,
		JsonRPC: "2.0",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error marshaling payload: %v", err)
	}

	url := "http://localhost" + httpListenAddress + "/"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	fmt.Printf("Response: %s\n", body)
}
