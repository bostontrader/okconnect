// The purpose of this file is to hold items required to communicate with a bookwerx server.
// These items are also present in okcatbox and ought to be factored out.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bostontrader/okconnect/config"
	"github.com/gojektech/heimdall/httpclient"
	"io/ioutil"
	"net/http"
)

type AId struct {
	Id uint32 `json:"accounts-id"`
}

type LID struct {
	LastInsertID uint32
}

// Given a response object, read the body and return it as a string.  Deal with the error message if necessary.
func bodyString(resp *http.Response) string {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("bookwerx-api.go body_string :%v", err)
	}
	return string(body)
}

func createDistribution(client *httpclient.Client, accountId uint32, amt int64, exp int32, txid uint32, cfg config.Config) (did uint32, err error) {

	url1 := fmt.Sprintf("%s/distributions", cfg.BookwerxConfig.Server)
	url2 := fmt.Sprintf("apikey=%s&account_id=%d&amount=%d&amount_exp=%d&transaction_id=%d",
		cfg.BookwerxConfig.APIKey, accountId, amt, exp, txid)

	h := make(map[string][]string)
	h["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	resp, err := client.Post(url1, bytes.NewBuffer([]byte(url2)), h)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go createDistribution 1: %v", err)
		fmt.Println(s)
		return 0, errors.New(s)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go createDistribution 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, bodyString(resp))
		fmt.Println(s)
		return 0, errors.New(s)
	}

	var insert LID
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&insert)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go createDistribution 3: %v", err)
		fmt.Println(s)
		return 0, errors.New(s)
	}

	return insert.LastInsertID, nil
}

func createTransaction(client *httpclient.Client, time string, cfg config.Config) (txid uint32, err error) {

	url1 := fmt.Sprintf("%s/transactions", cfg.BookwerxConfig.Server)
	url2 := fmt.Sprintf("apikey=%s&notes=deposit&time=%s", cfg.BookwerxConfig.APIKey, time)

	h := make(map[string][]string)
	h["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	resp, err := client.Post(url1, bytes.NewBuffer([]byte(url2)), h)
	defer resp.Body.Close()
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go createTransaction 1: %v", err)
		fmt.Println(s)
		return 0, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go createTransaction 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, bodyString(resp))
		fmt.Println(s)
		return 0, errors.New(s)
	}

	var insert LID
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&insert)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go createTransaction 3: %v", err)
		fmt.Println(s)
		return 0, errors.New(s)
	}
	txid = insert.LastInsertID

	return txid, nil
}
