package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"sbminimalistics.com/cctv-visionfusion/downloader/internal/config"
)

type RequestPacket struct {
	Token string `json:"token"`
}

type ResponsePacket struct {
	PictureURL              string `json:"picture_url"`
	DestinationPath         string `json:"destination_path"`
	DestinationBaseFilename string `json:"destination_base_filename"`
}

func ExchangeToken(apiEndpoint string, token string) (ResponsePacket, error) {
	url, err := url.JoinPath(config.Config.BackendBaseURL, apiEndpoint)
	fmt.Println("ExchangeToken url:", url)

	outPacket := RequestPacket{
		Token: token,
	}

	contents, err := json.Marshal(&outPacket)

	req, err := http.NewRequest("POST", url, bytes.NewReader(contents))

	req.Header.Add("Content-Type", "application/json")

	var client = &http.Client{
		Transport: &http.Transport{},
	}
	resp, err := client.Do(req)
	if err != nil {
		// log.Println("token exchange error:", err)
		fmt.Println("token exchange error:", err)
		return ResponsePacket{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("ExchangeToken token exchange response error:", err)
		return ResponsePacket{}, err
	}

	// fmt.Println("ExchangeToken body:", body)
	var inPacket ResponsePacket
	err = json.Unmarshal(body, &inPacket)
	if err != nil {
		log.Println("ExchangeToken unmarshal data error:", err)
		return ResponsePacket{}, err
	}

	fmt.Println("ExchangeToken token exchanged to:", string(body))

	return inPacket, nil
}
