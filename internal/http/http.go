package http

import (
	"fmt"
	"net/http"
	"io/ioutil"
)

func GetData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed %w", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read body %w", err)
	}

	return data, nil
}


