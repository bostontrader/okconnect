package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	bwapi "github.com/bostontrader/bookwerx-common-go"
	utils "github.com/bostontrader/okcommon"
	"github.com/bostontrader/okconnect/config"
	okchttp "github.com/bostontrader/okconnect/http"
	"github.com/gojektech/heimdall/httpclient"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type AccountTransferResult struct {
	TransferID     string `json:"transfer_id"`
	CurrencySymbol string `json:"currency"`
	From           string
	Amount         string
	To             string
	Result         string
}

// The bookwerx-core server will on occasion return JSON names that contain a '.'.  This vile habit
// causes trouble here.
// A good, bad, or ugly hack is to simply change the . to a -.  Do that here.
func fixDot(b []byte) {
	for i, num := range b {
		if num == 46 { // .
			b[i] = 45 // -
		}
	}
}

// The purpose of this function is to make a transfer between two different locations on OKEx (such as funding to spot)
// and to also create a transaction in the user's bookwerx to reflect said transfer.
//
// This function presently only supports transfers between 1 (spot) and 6 (funding).
// Example:
// okconnect transfer -currency BTC -quan 1.25 -from 6 -to 1 -config okconnect.yaml
// means transfer 1.25 BTC from funding to spot
//
// This function presently has a handful of important issues to be aware of.
//
// 1. Generally, in order to make this transfer, this function will make several API calls to OKEx and bookwerx that will
// change the state of each.  Each call has several ways to fail and unless everything works as h/o/p/e/d/ expected OKEx and bookwerx
// will not agree with each other.  There's no practical way to implement "transactioning" for this process so in the event of
// failure the user will just have to examine the state of OKEx _and_ bookwerx and cleanup whatever mess has been caused by
// the failure.
//
// 2. A fruitful source of error would be for the user to specify source and destinations that aren't properly
// configured in bookwerx.  When this function tries to create a bookwerx transaction it must determine actual account
// ids to dr and cr.  Either one of these may be absent and the transaction thus cannot be made.  It's tempting to try
// to merely create a new account, properly configured, to cure this woe, but doing so presents more trouble.  So at this time
// we don't do this.  Make sure all the expected accounts are present first before you use this function.
//
// 3. We make the API call to OKEx first because we need info from the correct results in order to
// subsequent create the bookwerx transaction.
//
// 4. In the event of some error that leaves OKEx and bookwerx in a disagreeable state,
// remember to use okconnect compare.

func Transfer(cfg *config.Config, transferCurrency *string, transferFrom *string, transferTo *string, transferQuan *string) {

	methodName := "okconnect:transfer.go:Transfer"

	// 1. First make the API call to OKEx

	// 1.1 Read the credentials file for OKEx
	credentials, err := config.ReadCredentialsFile(cfg.OKExConfig.Credentials)
	if err != nil {
		fmt.Printf("transfer.go:Transfer: Cannot read the OKEx credentials file.\n")
		return
	}

	// 1.2 Make the Call!
	_, err = accountTransfer(*cfg, *credentials)
	if err != nil {
		fmt.Printf("transfer.go:Transfer OKEx API call failed.\n")
		return
	}

	// 2. Validate the source and destination code and determine the relevant bookwerx categories to use.
	var catSource uint32
	var catDest uint32

	// 2.1 They should not be the same.
	if *transferFrom == *transferTo {
		fmt.Printf("transfer.go: The source and destination of this transfer are the same. No can do.\n")
		return
	}

	// 2.2 Source...
	if *transferFrom == "1" {
		catSource = cfg.BookwerxConfig.CatSpotAvailable
	} else if *transferFrom == "6" {
		catSource = cfg.BookwerxConfig.CatFunding
	} else {
		fmt.Printf("transfer.go: The transferFrom parameter %d must be 1 or 6\n", transferFrom)
		return
	}

	// 2.3 Destination...
	if *transferTo == "1" {
		catDest = cfg.BookwerxConfig.CatSpotAvailable
	} else if *transferTo == "6" {
		catDest = cfg.BookwerxConfig.CatFunding
	} else {
		fmt.Printf("transfer.go: The transferTo parameter %d must be 1 or 6\n", transferTo)
		return
	}

	// 3. Parse the quantity
	quan, err := decimal.NewFromString(*transferQuan)
	if err != nil {
		fmt.Printf("transfer.go: Cannot parse the quantity.\n")
		return
	}
	quanCoff := quan.Coefficient().Int64()

	// why two clients!? clean this up.
	// We'll need an HTTP client for the subsequent requests.
	timeout := 5000 * time.Millisecond
	clientB := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	// 4. Verify that the user has said currency defined.  If not, then the user _cannot_ have
	// a source account using said currency.

	// Building the url is rather tedious generally because of the need to escape
	// various parts of it.  More particularly:
	// http.client Requests cannot have spaces so we must use %20 instead.
	// the query string cannot have an = sign so we must use %3d instead.
	// fmt.Sprintf chokes on the % character so we must use %% instead.
	/*query := fmt.Sprintf("SELECT%%20currencies.id%%20FROM%%20currencies%%20WHERE%%20currencies.symbol%%3d'%s'", *transferCurrency)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.BookwerxConfig.Server, query, cfg.BookwerxConfig.APIKey)

	body, err := get(clientB, url)
	if err != nil {
		fmt.Printf("transfer.go: get request error: %v", err)
		return
	}
	fixDot(body)

	n := make([]CId, 0)
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&n)
	if err != nil {
		fmt.Printf("transfer.go: JSON decode error: %v", err)
		return;
	}

	if len(n) == 0 {
		fmt.Printf("transfer.go: Currency %s is not defined in bookwerx", "BTC")
		return
	} else if len(n) > 1 {
		fmt.Printf("transfer.go: There are more than one suitable currencies.  This should never happen.")
		return
	}*/

	// 5. Find the user's source account in his bookwerx db.  It's an account that is:
	// A. Tagged with the whatever category corresponds with the specified source, such as funding or spot,
	// B. Configured to use the specified currency.

	selectt := "SELECT%20accounts.id"
	from := "FROM%20accounts_categories"
	join1 := "JOIN%20accounts%20ON%20accounts.id%3daccounts_categories.account_id"
	join2 := "JOIN%20currencies%20ON%20currencies.id%3daccounts.currency_id"
	where := fmt.Sprintf("WHERE%%20category_id%%3d%d%%20AND%%20currencies.symbol%%3d'%s'", catSource, *transferCurrency)
	query := fmt.Sprintf("%s%%20%s%%20%s%%20%s%%20%s", selectt, from, join1, join2, where)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.BookwerxConfig.BaseURL, query, cfg.BookwerxConfig.APIKey)

	body, err := bwapi.Get(clientB, url)
	if err != nil {
		fmt.Printf("%s :Error reading %s\n%v", methodName, url, err)
		return
	}
	fixDot(body)

	n1 := make([]AId, 0)
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&n1)
	if err != nil {
		fmt.Printf("transfer.go: JSON decode error: %v\n", err)
		return
	}

	if len(n1) == 0 {
		fmt.Printf("transfer.go: Bookwerx does not have any account properly configured.\n")
		return
	} else if len(n1) > 1 {
		fmt.Printf("transfer.go: Bookwerx has more than one suitable account.  This should never happen.\n")
	}

	sourceAcctID := n1[0]

	// 6. Find the user's destination account in his bookwerx db in a manner similar to that of the source
	// account. It's an account that is:
	// A. Tagged with the whatever category corresponds with the specified destination, such as funding or spot,
	// B. Configured to use the specified currency.
	// Said destination account may not exist so create it if necessary.
	//selectt := "SELECT%20accounts.id";
	//from := "FROM%20accounts_categories";
	//join1 := "JOIN%20accounts%20ON%20accounts.id%3daccounts_categories.account_id";
	//join2 := "JOIN%20currencies%20ON%20currencies.id%3daccounts.currency_id";
	//cat1 := cfg.BookwerxConfig.FundingCat;
	//cat2 := cfg.BookwerxConfig.SpotAvailableCat;
	where = fmt.Sprintf("WHERE%%20category_id%%3d%d%%20AND%%20currencies.symbol%%3d'%s'", catDest, *transferCurrency)
	query = fmt.Sprintf("%s%%20%s%%20%s%%20%s%%20%s", selectt, from, join1, join2, where)
	url = fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.BookwerxConfig.BaseURL, query, cfg.BookwerxConfig.APIKey)

	body, err = bwapi.Get(clientB, url)
	if err != nil {
		fmt.Printf("transfer.go: get error: %v\n", err)
		return
	}

	// ugly hack to fix.  the '.' fucks up subsequent decoding
	for i, num := range body {
		if num == 46 {
			body[i] = 45
		}
	}

	n1 = make([]AId, 0)
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&n1)
	if err != nil {
		fmt.Printf("transfer.go: JSON decode error: %v\n", err)
		return
	}

	if len(n1) == 0 {
		fmt.Printf("transfer.go: Bookwerx does not have any account properly configured.\n")
		return
	} else if len(n1) > 1 {
		fmt.Printf("transfer.go: Bookwerx has more than one suitable account.  This should never happen.\n")
	}

	destAcctID := n1[0]

	/*where = fmt.Sprintf("WHERE accounts_categories.category_id %3d", catDest);
	url = fmt.Sprintf("%s %s %s %s %s %s %s", selectt, from, join1, join2, where, group, having);


	// 6.1 If there are zero accounts tagged with the suitable destination categories, then
	// create and tag them suitably
	if len(n.AccountId) == 0 {
		// A. Create the account and save the account id
		// B. Tag the account with destination & APIKEY8
	} else {
		// grab AccountId[0].  Assume that's the only one.
	}
	destAcctID := n.AccountId[0]
	*/

	// 7. Now execute the transfer on okex.
	// call okprobe to do the dirty work

	// 8. If successful, make the bookwerx tx.
	// 8. Now create the transaction on the user's books and the two distributions using three requests.

	// 8.1 Create the tx
	txid, err := createTransaction(clientB, "time", *cfg)
	if err != nil {
		//log.Error(err)
		//fmt.Fprintf(w, err.Error())
		fmt.Printf("transfer.go: Error creating bookwerx transaction.\n")
		return
	}

	// 8.2 Create the DR distribution
	_, err = createDistribution(clientB, destAcctID.Id, quanCoff, quan.Exponent(), txid, *cfg)
	if err != nil {
		//log.Error(err)
		//fmt.Fprintf(w, err.Error())
		fmt.Printf("transfer.go: Error creating bookwerx distribution.\n")
		return
	}

	// 8.3 Create the CR distribution
	_, err = createDistribution(clientB, sourceAcctID.Id, -quanCoff, quan.Exponent(), txid, *cfg)
	if err != nil {
		//log.Error(err)
		//fmt.Fprintf(w, err.Error())
		fmt.Printf("transfer.go: Error creating bookwerx distribution.\n")
		return
	}
	//fmt.Printf(catSource, catDest, quanCoff, sourceAcctID, destAcctID)
	return

}

// Make the API call to perform the transfer on okex
func accountTransfer(cfg config.Config, credentials utils.Credentials) (AccountTransferResult, error) {
	urlBase := cfg.OKExConfig.BaseURL
	endpoint := "/api/account/v3/transfer"
	url := urlBase + endpoint
	client := okchttp.GetHTTPClient(urlBase)
	reqBody := `{"from":"6", "to":"1", "amount":"0.1", "currency":"BTC"}`

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	prehash := timestamp + "POST" + endpoint + reqBody
	encoded, _ := utils.HmacSha256Base64Signer(prehash, credentials.SecretKey)

	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	if err != nil {
		fmt.Printf("transfer.go:accountTransfer: NewRequest error:%v\n", err)
		return AccountTransferResult{}, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("OK-ACCESS-KEY", credentials.Key)
	req.Header.Add("OK-ACCESS-SIGN", encoded)
	req.Header.Add("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Add("OK-ACCESS-PASSPHRASE", credentials.Passphrase)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("transfer.go:accountTransfer: client.Do error:%v\n", err)
		return AccountTransferResult{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("transfer.go:accountTransfer: ReadAll error:%v\n", err)
		return AccountTransferResult{}, err
	}

	err = resp.Body.Close()
	if err != nil {
		fmt.Printf("transfer.go:accountTransfer: Body.Close error:%v\n", err)
		return AccountTransferResult{}, err
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Status Code error: expected= 200, received=%d\n", resp.StatusCode)
		fmt.Printf("body=%s\n", string(respBody))
		return AccountTransferResult{}, errors.New("transfer.go:accountTransfer: status code error")
	}

	accountTransferResult := AccountTransferResult{}
	dec := json.NewDecoder(bytes.NewReader(respBody))
	err = dec.Decode(&accountTransferResult)
	if err != nil {
		fmt.Printf("transfer.go:accountTransfer: Wallet JSON decode error: %v\n", err)
		fmt.Printf("body=%s\n", string(respBody))
		return AccountTransferResult{}, err
	}

	return accountTransferResult, nil
}
