package openai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var apiKey string

func init() {
	apiKey = os.Getenv("OPEN_API_KEY")
	if apiKey == "" {
		log.Fatal("OPEN_API_KEY not set")
	}
}

var openaiURL = "https://api.openai.com/v1/"

func url(relativeAddress string) string {
	return openaiURL + relativeAddress
}

type ImageRequest struct {
	Prompt         string `json:"prompt"`
	Size           string `json:"size"`
	ResponseFormat string `json:"response_format"`
}

type ImageResponse struct {
	Created int64               `json:"created"`
	Data    []ImageResponseData `json:"data"`
}

type ImageResponseData struct {
	B64JSON string `json:"b64_json"`
}

func GenerateImage(prompt string) ([]byte, error) {

	url := url("images/generations")
	request := ImageRequest{
		Prompt:         prompt,
		Size:           "1024x1024",
		ResponseFormat: "b64_json",
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := callOpenAI(url, requestBody)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result ImageResponse
	if err := json.Unmarshal([]byte(responseBody), &result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("No data returned")
	}
	decodedData, err := base64.StdEncoding.DecodeString(result.Data[0].B64JSON)
	if err != nil {
		return nil, err
	}
	return decodedData, nil

}

func callOpenAI(url string, payload []byte) (*http.Response, error) {

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{
		Transport: &LoggingTransport{Transport: http.DefaultTransport},
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

type LoggingTransport struct {
	Transport http.RoundTripper
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request body to a buffer
	var requestBodyCopy bytes.Buffer
	_, err := io.Copy(&requestBodyCopy, req.Body)
	if err != nil {
		return nil, err
	}
	req.Body.Close()
	req.Body = io.NopCloser(&requestBodyCopy)

	// Log the request body (copy)
	fmt.Println("Request Body:", requestBodyCopy.String())

	// Make the actual HTTP request
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Clone the response body to a buffer
	var responseBodyCopy bytes.Buffer
	_, err = io.Copy(&responseBodyCopy, resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(&responseBodyCopy)

	// Log the response body (copy)
	fmt.Println("Response Body:", responseBodyCopy.String())

	return resp, nil
}
