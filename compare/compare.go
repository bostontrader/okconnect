package compare

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"github.com/bostontrader/okconnect/config"
	okchttp "github.com/bostontrader/okconnect/http"
	"github.com/shopspring/decimal"
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

// 1.2. Define a Comparison struct that will enable us to assemble the bits 'n' pieces of information we find about these balances.
type Comparison struct {
	Category        string // Funding, Spot-Hold, etc.
	OKExBalance     MaybeBalance
	BookwerxBalance MaybeBalance
	CurrencySymbol  string // OKEx uses a currency symbol as a currency id
	AccountID       uint32 // This is the account id for bookwerx
}

// 2. Define some structs for use in interfacing with Bookwerx
type AccountCurrency struct {
	AccountID uint32 `json:"account_id"`
	Title     string
	Currency  CurrencySymbol
}

type BalanceResultDecorated struct {
	Account AccountCurrency
	Sum     DFP
}

type CurrencySymbol struct {
	CurrencyID uint32 `json:"currency_id"`
	Symbol     string
}

type DFP struct {
	Amount int64
	Exp    int8
}

type Sums struct {
	Sums []BalanceResultDecorated
}

// Make the API call to get all funding balances from OKEx
func getWallet(cfg config.Config, credentials utils.Credentials) ([]utils.WalletEntry, error) {
	urlBase := cfg.OKExConfig.Server
	endpoint := "/api/account/v3/wallet"
	url := urlBase + endpoint
	client := okchttp.GetHTTPClient(urlBase)
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	prehash := timestamp + "GET" + endpoint
	encoded, _ := utils.HmacSha256Base64Signer(prehash, credentials.SecretKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("NewRequest error: %v\n", err)
		return nil, err
	}

	req.Header.Add("OK-ACCESS-KEY", credentials.Key)
	req.Header.Add("OK-ACCESS-SIGN", encoded)
	req.Header.Add("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Add("OK-ACCESS-PASSPHRASE", credentials.Passphrase)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("client.Do error: %v\n", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ReadAll error: %v\n", err)
		return nil, err
	}

	_ = resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Status Code error: expected= 200, received=%d\n", resp.StatusCode)
		fmt.Printf("body=%s\n", string(body))
		return nil, errors.New("status code error")
	}

	walletEntries := make([]utils.WalletEntry, 0)
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	err = dec.Decode(&walletEntries)
	if err != nil {
		fmt.Printf("Wallet JSON decode error: %v\n", err)
		return nil, err
	}

	return walletEntries, nil
}

// Make the API call to get all spot balances from OKEx.  This gives us both available and hold balances.
func getAccounts(cfg config.Config, credentials utils.Credentials) ([]utils.AccountsEntry, error) {
	urlBase := cfg.OKExConfig.Server
	endpoint := "/api/spot/v3/accounts"
	url := urlBase + endpoint
	client := okchttp.GetHTTPClient(urlBase)
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	prehash := timestamp + "GET" + endpoint
	encoded, _ := utils.HmacSha256Base64Signer(prehash, credentials.SecretKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("NewRequest error: %v\n", err)
		return nil, err
	}

	req.Header.Add("OK-ACCESS-KEY", credentials.Key)
	req.Header.Add("OK-ACCESS-SIGN", encoded)
	req.Header.Add("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Add("OK-ACCESS-PASSPHRASE", credentials.Passphrase)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("client.Do error: %v\n", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ReadAll error: %v\n", err)
		return nil, err
	}

	_ = resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Status Code error: expected= 200, received=%d\n", resp.StatusCode)
		fmt.Printf("body=%s\n", string(body))
		return nil, errors.New("status code error")
	}

	accountsEntries := make([]utils.AccountsEntry, 0)
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	err = dec.Decode(&accountsEntries)
	if err != nil {
		fmt.Printf("Accounts JSON decode error: %v\n", err)
		return nil, err
	}

	return accountsEntries, nil
}

// Get the current balances of all accounts tagged with a list of categories from Bookwerx.
func getCategoryDistSums(url string) ([]BalanceResultDecorated, error) {

	client := okchttp.GetHTTPClient(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("NewRequest error: %v\n", err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("client.Do error: %v\n", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ReadAll error: %v\n", err)
		return nil, err
	}

	_ = resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Status Code error: expected= 200, received=%d\n", resp.StatusCode)
		fmt.Printf("body=%s\n", string(body))
		return nil, errors.New("status code error")
	}

	n := Sums{}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&n)
	if err != nil {
		fmt.Printf("getCategoryDistSums JSON decode error: %v\n", err)
		return nil, err
	}

	return n.Sums, nil
}

func Compare(cfg *config.Config) {

	// 1. Read the credentials file for OKEx
	credentials, err := config.ReadCredentialsFile(cfg.OKExConfig.Credentials)
	if err != nil {
		fmt.Printf("Cannot read the OKEx credentials file.\n")
		return
	}

	// 2. Get the funding balances

	// 2.1 ... from OKEx
	walletEntries, err := getWallet(*cfg, *credentials)
	if err != nil {
		fmt.Printf("Cannot execute the wallet API endpoint.\n")
		return
	}

	// 2.1.1 Init the comparison chart for the funding section
	comparisonEntriesFunding := make(map[string]Comparison)
	for _, walletEntry := range walletEntries {

		b, err := decimal.NewFromString(walletEntry.Balance)
		mb := MaybeBalance{b, false}
		if err != nil {
			mb = MaybeBalance{decimal.NewFromInt(0), true}
		}

		comparison := Comparison{
			"F",
			mb,
			MaybeBalance{decimal.NewFromInt(0), true},
			walletEntry.CurrencyID,
			0,
		}
		comparisonEntriesFunding[walletEntry.CurrencyID] = comparison // this is really the currency symbol
	}

	// 2.2 ... from Bookwerx
	// Get the account balances for all accounts tagged as funding_cat.
	categories := fmt.Sprintf("%d", cfg.BookwerxConfig.FundingCat)
	url := fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.BookwerxConfig.Server, cfg.BookwerxConfig.APIKey, categories)

	sums, err := getCategoryDistSums(url)
	if err != nil {
		fmt.Printf("Cannot execute the getCategoryDistSums API endpoint.\n")
		return
	}

	// 2.2.1. Insert whatever balance info is found into the comparison chart for the funding section.  Modify an existing record or create a new one if necessary.
	for _, brd := range sums {

		b1 := decimal.New(brd.Sum.Amount, int32(brd.Sum.Exp))

		i, ok := comparisonEntriesFunding[brd.Account.Currency.Symbol]
		if ok {
			// The entry is found, replace the BookwerxBalance
			i.BookwerxBalance = MaybeBalance{b1, false}
			i.AccountID = brd.Account.AccountID
			comparisonEntriesFunding[brd.Account.Currency.Symbol] = i
		} else {
			// The entry is not found, build a new entry
			comparisonEntriesFunding[brd.Account.Currency.Symbol] = Comparison{
				"F",
				MaybeBalance{decimal.NewFromInt(0), true},
				MaybeBalance{b1, false},
				brd.Account.Currency.Symbol,
				brd.Account.AccountID,
			}
		}
	}

	// 3. Get the spot balances.  Be aware of available and hold balances.

	// 3.1 ... from OKEx
	accountsEntries, err := getAccounts(*cfg, *credentials)
	if err != nil {
		fmt.Printf("Cannot execute the accounts API endpoint.\n")
		return
	}

	// 3.1.1 Init the comparison chart for the spot, available section
	comparisonEntriesSpotA := make(map[string]Comparison)
	for _, accountsEntry := range accountsEntries {

		b, err := decimal.NewFromString(accountsEntry.Balance)
		mb := MaybeBalance{b, false}
		if err != nil {
			mb = MaybeBalance{decimal.NewFromInt(0), true}
		}

		comparison := Comparison{
			"Spot-Available",
			mb,
			MaybeBalance{decimal.NewFromInt(0), true},
			accountsEntry.CurrencyID, // okex uses a currency symbol as their currency id
			0,
		}
		comparisonEntriesSpotA[accountsEntry.CurrencyID] = comparison // this is really the currency symbol
	}

	// 3.1.2 Init the comparison chart for the spot, hold section
	comparisonEntriesSpotH := make(map[string]Comparison)
	for _, accountsEntry := range accountsEntries {

		b, err := decimal.NewFromString(accountsEntry.Balance)
		mb := MaybeBalance{b, false}
		if err != nil {
			mb = MaybeBalance{decimal.NewFromInt(0), true}
		}

		comparison := Comparison{
			"Spot-Hold",
			mb,
			MaybeBalance{decimal.NewFromInt(0), true},
			accountsEntry.CurrencyID, // okex uses a currency symbol as their currency id
			0,
		}
		comparisonEntriesSpotH[accountsEntry.CurrencyID] = comparison // this is really the currency symbol
	}

	// 3.2 ... from Bookwerx

	// 3.2.1 ... for accounts tagged as spot_available_cat.

	// 3.2.1.1 Get the account balances for all relevant accounts.
	/*categories = fmt.Sprintf("%d", cfg.BookwerxConfig.FundingCat)
	url = fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.BookwerxConfig.Server, cfg.BookwerxConfig.APIKey, categories)

	sums, err = getCategoryDistSums(url)
	if err != nil {
		fmt.Printf("Cannot execute the getCategoryDistSums API endpoint.")
		return
	} */

	// 3.2.1.2 Insert whatever balance info is found into the comparison chart for the spot-available section.
	/*for _, brd := range sums {

			b1 := decimal.New(brd.Sum.Amount, uint32(brd.Sum.Exp))

			i, ok := comparisonEntriesFunding[brd.Account.Currency.Symbol]
			if ok {
				// The entry is found, replace the BookwerxBalance
				i.BookwerxBalance = MaybeBalance{b1, false}
				i.AccountID = brd.Account.AccountID
				comparisonEntriesSpotH[brd.Account.Currency.Symbol] = i
			} else {
				// The entry is not found, build a new entry
				comparisonEntriesSpotH[brd.Account.Currency.Symbol] = Comparison{
					"Spot-Available",
	        		MaybeBalance{decimal.NewFromInt(0), true},
					MaybeBalance{b1, false},
					brd.Account.Currency.Symbol,
					brd.Account.AccountID,
				}
			}
		} */

	// 3.2.2 ... for accounts tagged as spot_hold_cat.

	// 3.2.2.1 Get the account balances for all relevant accounts.
	/*categories = fmt.Sprintf("%d", cfg.BookwerxConfig.FundingCat)
	url = fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.BookwerxConfig.Server, cfg.BookwerxConfig.APIKey, categories)

	sums, err = getCategoryDistSums(url)
	if err != nil {
		fmt.Printf("Cannot execute the getCategoryDistSums API endpoint.")
		return
	}

	// 3.2.2.2 Insert whatever balance info is found into the comparison chart for the funding section.
	for _, brd := range sums {

		b1 := decimal.New(brd.Sum.Amount, uint32(brd.Sum.Exp))

		i, ok := comparisonEntriesSpotH[brd.Account.Currency.Symbol]
		if ok {
			// The entry is found, replace the BookwerxBalance
			i.BookwerxBalance = MaybeBalance{b1, false}
			comparisonEntriesSpotH[brd.Account.Currency.Symbol] = i
		} else {
			// The entry is not found, build a new entry
			comparisonEntriesSpotH[brd.Account.Currency.Symbol] = Comparison{
				OKExBalance:     MaybeBalance{decimal.NewFromInt(0), true},
				BookwerxBalance: MaybeBalance{b1, false},
			}
		}
	} */

	// 4. Build the return value
	retValA := make([]Comparison, 0)

	// 4.1 Funding
	for _, v := range comparisonEntriesFunding {
		b1, b2 := decimal.RescalePair(v.BookwerxBalance.Balance, v.OKExBalance.Balance)
		if !b1.Equal(b2) {
			retValA = append(retValA, v)
		}
	}

	// 4.2 Spot

	// 4.2.1 Spot Available
	for _, v := range comparisonEntriesSpotA {
		b1, b2 := decimal.RescalePair(v.BookwerxBalance.Balance, v.OKExBalance.Balance)
		if !b1.Equal(b2) {
			retValA = append(retValA, v)
		}
	}

	// 4.2.2 Spot Hold
	for _, v := range comparisonEntriesSpotH {
		b1, b2 := decimal.RescalePair(v.BookwerxBalance.Balance, v.OKExBalance.Balance)
		if !b1.Equal(b2) {
			retValA = append(retValA, v)
		}
	}

	retValB, _ := json.Marshal(retValA)

	fmt.Printf(string(retValB))

}
