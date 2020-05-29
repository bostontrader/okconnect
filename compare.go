package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"time"
)

/*
As we try to make the comparison between the OKEx account balances and the same in Bookwerx we have a problem whereby an account may exist on one side but not the other.  This problem is further compounded because we have a configuration that specifies the mapping between OKEx and Bookwerx accounts.  This configuration may be missing some of these accounts or contain extraneous references.  How shall we untangle this?

A solution is as follows:

1. By examining the ComparisonConfig object, the OKEx API, and the Bookwerx API, we will find many balances.  Our general goal is to organize this mass of information into a format amenable to analysis and debuggery.
*/

// 2. Define a MaybeBalance struct that simulates a Maybe.
type MaybeBalance struct {
	Balance decimal.Decimal
	Nil     bool // Is the balance really supposed to be nil?
}

// 3. Define a Comparison struct that will enable us to assemble the bits 'n' pieces of information we find about these balances. Instances of this struct are expected to live in a collection that is indexed using a CurrencyID so we don't need that in the struct.
type Comparison struct {
	Config            bool // Did we find this Comparison by examining the config?
	BookwerxAccountID int
	OKExBalance       MaybeBalance
	BookwerxBalance   MaybeBalance
}

func GetClient(urlBase string) (client *http.Client) {

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

func getCredentials(keyFile string) utils.Credentials {
	var obj utils.Credentials
	data, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return obj
}

func Compare(config Config) {

	// 1. Examine the funding accounts

	comparisonEntries := make(map[string]Comparison)

	// 1.1 Make the API call to get all funding balances
	urlBase := config.OKExConfig.Server
	endpoint := "/api/account/v3/wallet"
	url := urlBase + endpoint
	client := GetClient(urlBase)
	credentials := getCredentials(config.OKExConfig.Credentials)
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	prehash := timestamp + "GET" + endpoint
	encoded, _ := utils.HmacSha256Base64Signer(prehash, credentials.SecretKey)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("OK-ACCESS-KEY", credentials.Key)
	req.Header.Add("OK-ACCESS-SIGN", encoded)
	req.Header.Add("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Add("OK-ACCESS-PASSPHRASE", credentials.Passphrase)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Read the body into a []byte and then create a new io.Reader using this []byte.  This enables us to close resp.Body, which we must do, and create an io.Reader which we will need in order to Decode JSON.
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("error:\nexpected= ", 200, "\nreceived=", resp.StatusCode)
		log.Printf("body=", string(body))
	}

	walletEntries := make([]utils.WalletEntry, 0)
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	err = dec.Decode(&walletEntries)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// 1.2 Examine every entry in Compare config.  Insert each of them into comparisonEntries, config=true, no balances.  If there are duplicate currency ids configured, they will added/replaced in the order found with at most one entry saved.
	for _, configEntry := range config.CompareConfig.Funding {
		comparisonEntries[configEntry.CurrencyID] = Comparison{
			Config:            true,
			BookwerxAccountID: configEntry.Available,
			OKExBalance:       MaybeBalance{decimal.New(0, 0), true},
			BookwerxBalance:   MaybeBalance{decimal.New(0, 0), true},
		}
	}

	// 1.3 Examine every entry in the wallet.  Assume this contains no duplicated currency ids. If a corresponding entry exists in comparisonEntries then modify its OKEx balance, else create a new entry, config=false, with the OKEx balance.
	for _, walletEntry := range walletEntries {

		// 1.3.1 Parse the walletEntry balance
		b, err := decimal.NewFromString(walletEntry.Available)
		var mb = MaybeBalance{}
		if err != nil {
			mb.Balance = b
		} else {
			mb.Nil = true
		}

		// 1.3.2 Search for the walletEntry's currency in comparisonEntries.
		i, ok := comparisonEntries[walletEntry.CurrencyID]
		if ok {
			// The entry is found, replace the OKExBalance
			i.OKExBalance = mb
			comparisonEntries[walletEntry.CurrencyID] = i
		} else {
			// The entry is not found, build a new entry
			comparisonEntries[walletEntry.CurrencyID] = Comparison{
				Config:            false,
				BookwerxAccountID: -1, // means that we don't have one
				OKExBalance:       mb,
				BookwerxBalance:   MaybeBalance{decimal.New(0, 0), true},
			}
		}
	}

	// 1.4 Examine every entry of comparisonEntries where config=true.  Request distributions for that account from Bookwerx.
	// If the API call succeeds, compute the balance and set it in the comparisonEntry, else said account does not exist so leave the comparisonEntry balance nil.
	for _, comparisonEntry := range comparisonEntries {
		if comparisonEntry.Config {
			// 1.4.1 Make the API call to Bookwerx to get the balance information for the specified account.
			url = fmt.Sprintf("%s/distributions/for_account?apikey=%s&account_id=%d",
				config.BookwerxConfig.Server, config.BookwerxConfig.APIKey,
				comparisonEntry.BookwerxAccountID)

			req, err = http.NewRequest("GET", url, nil)
			resp, err = client.Do(req)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			// 1.4.2 Read the body into a []byte and then create a new io.Reader using this []byte.  This enables us to close resp.Body, which we must do, and create an io.Reader which the caller needs in order to Decode JSON.
			body, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			if resp.StatusCode != 200 {
				log.Printf("error:\nexpected= ", 200, "\nreceived=", resp.StatusCode)
				log.Printf("body=", string(body))
			}

			// 1.4.3 Parse the body into an array of DistributionJoinedEntry
			distributionJoinedEntries := make([]DistributionJoinedEntry, 0)
			dec = json.NewDecoder(bytes.NewReader(body))
			dec.DisallowUnknownFields()
			err = dec.Decode(&distributionJoinedEntries)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			// 1.4.4 Calculate the sum of the distributions
			//sum := DFP{0, 0}
			sum := decimal.Decimal{}
			for _, e := range distributionJoinedEntries {
				sum = sum.Add(decimal.New(e.Amount, int32(e.AmountExp)))
			}
			comparisonEntry.BookwerxBalance = MaybeBalance{sum, true}
		}
	}

	// 1.5 We don't need to examine in Bookwerx in order to identify accounts that should be in OKEx because the prior gyrations are sufficient to figure all this out.  If an OKEx balance exists it will be seen in the API.  If an OKEx balance does not exist, then there's no harm in Bookwerx not knowing about it.

	// 1.6 Now it's time for the analysis.
	for i, comparisonEntry := range comparisonEntries {
		fmt.Println("Currency ID:", i, ", In config:", comparisonEntry.Config, ", OKEx balance:", comparisonEntry.OKExBalance, ", Bookwerx balance:", comparisonEntry.BookwerxBalance)

		if comparisonEntry.Config == true {
			if reflect.DeepEqual(comparisonEntry.OKExBalance, comparisonEntry.BookwerxBalance) {
				// The balances are the same.  Cool.
			} else {
				fmt.Println("The balances for Currency ID", i, "do not match.")
			}
		} else {
			fmt.Println("Currency ID ", i, "was found in an OKEx API call but is not present in the configuration.")
		}

	}

}

type DistributionJoinedEntry struct {
	AccountTitle     string `json:"account_title"`
	AccountID        int    `json:"aid"`
	Amount           int64  `json:"amount"`
	AmountExp        int8   `json:"amount_exp"`
	APIKey           string `json:"apikey"`
	ID               int    `json:"id"`
	TransactionID    int    `json:"tid"`
	TransactionNotes string `json:"tx_notes"`
	TransactionTime  string `json:"tx_time"`
}
