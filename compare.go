package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

// 1. Define some structs used for the comparison of balances.

// In making the comparison between the OKEx account balances and the same in Bookwerx we have a general problem whereby an account may exist on one side but not the other.
// A solution is as follows:

// 1.1 A MaybeBalance simulates a Maybe.
type MaybeBalance struct {
	Balance decimal.Decimal
	Nil     bool // Is the balance really supposed to be nil?
}

// 1.2. Define a Comparison struct that will enable us to assemble the bits 'n' pieces of information we find about these balances. Instances of this struct are expected to live in a collection that is indexed using a currency symbol so we don't need that in the struct.
type Comparison struct {
	//BookwerxAccountID int
	OKExBalance     MaybeBalance
	BookwerxBalance MaybeBalance
}

// 2. Define some structs for use in interfacing with Bookwerx
type AccountCurrency struct {
	AccountID int32
	Title     string
	Currency  CurrencySymbol
}

type BalanceResultDecorated struct {
	Account AccountCurrency
	Sum     DFP
}

type CurrencySymbol struct {
	CurrencyID int32
	Symbol     string
}

type DFP struct {
	Amount int64
	Exp    int8
}

type Sums struct {
	Sums []BalanceResultDecorated
}

func getHTTPClient(urlBase string) (client *http.Client) {

	if len(urlBase) >= 6 && urlBase[:6] == "https:" {
		tr := &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		}
		return &http.Client{Transport: tr}
	}

	return &http.Client{}

}

func readCredentialsFile(keyFile string) (utils.Credentials, error) {
	var obj utils.Credentials
	data, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatalf("error: %v", err)
		return utils.Credentials{}, err
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		log.Fatalf("error: %v", err)
		return utils.Credentials{}, err
	}
	return obj, nil
}

// Make the API call to get all funding balances from OKEx
func getWallet(cfg Config, credentials utils.Credentials) ([]utils.WalletEntry, error) {
	urlBase := cfg.OKExConfig.Server
	endpoint := "/api/account/v3/wallet"
	url := urlBase + endpoint
	client := getHTTPClient(urlBase)
	//credentials := readCredentialsFile(cfg.OKExConfig.Credentials)
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	prehash := timestamp + "GET" + endpoint
	encoded, _ := utils.HmacSha256Base64Signer(prehash, credentials.SecretKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("error: %v", err)
		return nil, err
	}

	req.Header.Add("OK-ACCESS-KEY", credentials.Key)
	req.Header.Add("OK-ACCESS-SIGN", encoded)
	req.Header.Add("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Add("OK-ACCESS-PASSPHRASE", credentials.Passphrase)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error: %v", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error: %v", err)
		return nil, err
	}

	_ = resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Status Code error: expected= 200, received=", resp.StatusCode)
		log.Fatalf("body=", string(body))
		return nil, errors.New("Status code error")
	}

	walletEntries := make([]utils.WalletEntry, 0)
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	err = dec.Decode(&walletEntries)
	if err != nil {
		log.Fatalf("error: %v", err)
		return nil, err
	}

	return walletEntries, nil
}

// Make the API call to get all spot balances from OKEx.  This gives us both available and hold balances.
func getAccounts(cfg Config, credentials utils.Credentials) ([]utils.WalletEntry, error) {
	urlBase := cfg.OKExConfig.Server
	endpoint := "/api/account/v3/wallet"
	url := urlBase + endpoint
	client := getHTTPClient(urlBase)
	// credentials := readCredentialsFile(cfg.OKExConfig.Credentials)
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	prehash := timestamp + "GET" + endpoint
	encoded, _ := utils.HmacSha256Base64Signer(prehash, credentials.SecretKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("OK-ACCESS-KEY", credentials.Key)
	req.Header.Add("OK-ACCESS-SIGN", encoded)
	req.Header.Add("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Add("OK-ACCESS-PASSPHRASE", credentials.Passphrase)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	_ = resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("Status Code error: expected= 200, received=", resp.StatusCode)
		log.Println("body=", string(body))
		return nil, errors.New("Status code error")
	}

	walletEntries := make([]utils.WalletEntry, 0)
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	err = dec.Decode(&walletEntries)
	if err != nil {
		return nil, err
	}

	return walletEntries, nil
}

// Get the current balances of all accounts tagged with a list of categories from Bookwerx.
func getCategoryDistSums(url string) ([]BalanceResultDecorated, error) {

	client := getHTTPClient(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	_ = resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("Status Code error: expected= 200, received=", resp.StatusCode)
		log.Println("body=", string(body))
		return nil, errors.New("Status code error")
	}

	n := Sums{}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&n)
	if err != nil {
		return nil, err
	}

	return n.Sums, nil
}

func Compare(cfg Config) {

	// 1. Read the credentials file for OKEx
	credentials, err := readCredentialsFile(cfg.OKExConfig.Credentials)
	if err != nil {
		return
	}

	// 2. Get the funding balances

	// 2.1 ... from OKEx
	walletEntries, err := getWallet(cfg, credentials)
	if err != nil {
		return
	}

	// 2.1.1 Init the comparison chart for the funding section
	comparisonEntries := make(map[string]Comparison)
	for _, walletEntry := range walletEntries {

		b, err := decimal.NewFromString(walletEntry.Balance)
		mb := MaybeBalance{b, false}
		if err != nil {
			mb = MaybeBalance{decimal.NewFromInt(0), true}
		}

		comparison := Comparison{mb, MaybeBalance{decimal.NewFromInt(0), true}}
		comparisonEntries[walletEntry.CurrencyID] = comparison // this is really the currency symbol
	}

	// 2.2 ... from Bookwerx
	// Get the account balances for all accounts tagged as funding_cat.
	categories := fmt.Sprintf("%d", cfg.BookwerxConfig.FundingCat)
	url := fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.BookwerxConfig.Server, cfg.BookwerxConfig.APIKey, categories)

	sums, err := getCategoryDistSums(url)
	if err != nil {
		return
	}

	// 2.2.1. Insert whatever balance info is found into the comparison chart for the funding section.
	for _, brd := range sums {

		b1 := decimal.New(brd.Sum.Amount, int32(brd.Sum.Exp))

		i, ok := comparisonEntries[brd.Account.Currency.Symbol]
		if ok {
			// The entry is found, replace the BookwerxBalance
			i.BookwerxBalance = MaybeBalance{b1, false}
			comparisonEntries[brd.Account.Currency.Symbol] = i
		} else {
			// The entry is not found, build a new entry
			comparisonEntries[brd.Account.Currency.Symbol] = Comparison{
				OKExBalance:     MaybeBalance{decimal.NewFromInt(0), true},
				BookwerxBalance: MaybeBalance{b1, false},
			}
		}
	}

	// 3. Print the final analysis.
	errorCnt := 0
	for k, v := range comparisonEntries {

		b1, b2 := decimal.RescalePair(v.BookwerxBalance.Balance, v.OKExBalance.Balance)
		if !b1.Equal(b2) {
			log.Printf("%s OKEx=%v Bookwerx=%v", k, v.OKExBalance, v.BookwerxBalance)
		}
	}

	if errorCnt == 0 {
		log.Println("All relevant balances agree with each other.")
	} else {
		log.Println("There is at lease one balance error.")
	}
}
