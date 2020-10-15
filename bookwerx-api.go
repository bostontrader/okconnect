// The purpose of this file is to hold items required to communicate with a bookwerx server.
// These items are also present in okcatbox and ought to be factored out.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gojektech/heimdall/httpclient"
	"io/ioutil"
	"net/http"
)

type CId struct {
	Id int32 `json:"currencies-id"`
}

type AId struct {
	Id int32 `json:"accounts-id"`
}

type LID struct {
	LastInsertID int32
}

// Given a response object, read the body and return it as a string.  Deal with the error message if necessary.
func body_string(resp *http.Response) string {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("bookwerx-api.go body_string :%v", err)
	}
	return string(body)
}

func createDistribution(client *httpclient.Client, account_id int32, amt int64, exp int32, txid int32, cfg Config) (did int32, err error) {

	url1 := fmt.Sprintf("%s/distributions", cfg.BookwerxConfig.Server)
	url2 := fmt.Sprintf("apikey=%s&account_id=%d&amount=%d&amount_exp=%d&transaction_id=%d",
		cfg.BookwerxConfig.APIKey, account_id, amt, exp, txid)

	h := make(map[string][]string)
	h["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	resp, err := client.Post(url1, bytes.NewBuffer([]byte(url2)), h)
	defer resp.Body.Close()
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go createDistribution 1: %v", err)
		fmt.Println(s)
		return -1, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go createDistribution 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		fmt.Println(s)
		return -1, errors.New(s)
	}

	var insert LID
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&insert)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go createDistribution 3: %v", err)
		fmt.Println(s)
		return -1, errors.New(s)
	}

	return insert.LastInsertID, nil
}

func createTransaction(client *httpclient.Client, time string, cfg Config) (txid int32, err error) {

	url1 := fmt.Sprintf("%s/transactions", cfg.BookwerxConfig.Server)
	url2 := fmt.Sprintf("apikey=%s&notes=deposit&time=%s", cfg.BookwerxConfig.APIKey, time)

	h := make(map[string][]string)
	h["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	resp, err := client.Post(url1, bytes.NewBuffer([]byte(url2)), h)
	defer resp.Body.Close()
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go createTransaction 1: %v", err)
		fmt.Println(s)
		return -1, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go createTransaction 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		fmt.Println(s)
		return -1, errors.New(s)
	}

	var insert LID
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&insert)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go createTransaction 3: %v", err)
		fmt.Println(s)
		return -1, errors.New(s)
	}
	txid = insert.LastInsertID

	return txid, nil
}

// Generic get on the API
func get(client *httpclient.Client, url string) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("bookwerx-api get NewRequest error: %v", err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("bookwerx-api get client.Do error: %v", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("bookwerx-api get ReadAll error: %v", err)
		return nil, err
	}
	fmt.Println(string(body))
	_ = resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Status Code error: expected= 200, received=", resp.StatusCode)
		fmt.Println("body=", string(body))
		return nil, err
	}

	return body, nil

}
