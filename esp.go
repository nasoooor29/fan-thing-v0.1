package main

import (
	"net/http"
	"strconv"
	"strings"
)

func SendPostRequest(fanSpeed int) error {
	resp, err := http.Post(
		ESP_URL,
		"text/plain",
		strings.NewReader(strconv.Itoa(fanSpeed)),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
