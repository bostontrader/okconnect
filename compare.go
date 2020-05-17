package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"
)

/*
As we try to make the comparison between the OKEx account balances and the same in Bookwerx we have a problem whereby an account may exist on one side but not the other.  This problem is further compounded because we have a configuration that specifies the mapping between OKEx and Bookwerx accounts.  This configuration may also be missing some of these accounts so we must be aware of this also.

A solution is as follows:

1. By examining the ComparisonConfig object, the OKEx API, and the Bookwerx API, we will find many balances.  Our general goal is to organize this mass of information into a format amenable to analysis and debuggery.
*/

// 2. Define an enumeration of balance sources.  We are going to collect many balances and we will need their sources in order to perform the inevitable debuggery.
type BalanceSource int

const (
	CONFIG BalanceSource = 1 + iota
	BOOKWERX
	OKEX_FUNDING
	OKEX_SPOT
)

// 3. Define a Balance struct that will contain a Balance and a source.
type Balance struct {
	BalanceSource BalanceSource
	Balance       int
}

// 4. Define a Comparison struct that will contain a Balance for OKEx, Bookwerk, or ideally both of them.
type Comparison struct {
	OKExBalance     Balance
	BookwerxBalance Balance
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
		panic(err)
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		panic(err)
	}
	return obj
}

func Compare(config Config) {

	// 5. Starting with the CompareConfig, we can retrieve balances from both OKEx and Bookwerx.

	// 5.1 Get each balance in okex funding
	// First make the API call to get all balances
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
		panic(err)
	}

	if resp.StatusCode != 200 {
		fmt.Println("error:\nexpected= ", 200, "\nreceived=", resp.StatusCode)
	}

	// Read the body into a []byte and then create a return a new io.Reader using this []byte.  This enables us to close resp.Body, which we must do, and return an io.Reader which the caller needs in order to Decode JSON.
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		panic(err)
	}
	//return bytes.NewReader(body)
	body1 := bytes.NewReader(body)

	walletEntries := make([]utils.WalletEntry, 0)
	dec := json.NewDecoder(body1)
	dec.DisallowUnknownFields()
	err = dec.Decode(&walletEntries)
	if err != nil {
		panic(err)
	}
	fmt.Println(&walletEntries)
	fmt.Println(reflect.TypeOf(walletEntries))

	for _, e := range walletEntries {
		fmt.Println("e1=", e)
		fmt.Println("hold=", e.Hold)
		fmt.Println("hold=", e.Available)

		av := parse(e.Available)
		b1 := Balance{CONFIG, 666}
		fmt.Println("parse", av, b1)
		//for _, e2 := range e1.Currencies {
		//fmt.Println("e2=", e2)
		//}
	}

	url = "http://192.168.0.100:3003/distributions/for_account?apikey=catfood&account_id=314"
	req, err = http.NewRequest("GET", url, nil)
	resp, err = client.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		fmt.Println("error:\nexpected= ", 200, "\nreceived=", resp.StatusCode)
	}

	// Read the body into a []byte and then create a return a new io.Reader using this []byte.  This enables us to close resp.Body, which we must do, and return an io.Reader which the caller needs in order to Decode JSON.
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		panic(err)
	}
	//fmt.Println(string(body))
	body1 = bytes.NewReader(body)

	distributionJoinedEntries := make([]DistributionJoinedEntry, 0)
	dec = json.NewDecoder(body1)
	dec.DisallowUnknownFields()
	err = dec.Decode(&distributionJoinedEntries)
	if err != nil {
		panic(err)
	}
	fmt.Println(&distributionJoinedEntries)
	fmt.Println(reflect.TypeOf(distributionJoinedEntries))

	sum := DFP{0, 0}
	for _, e := range distributionJoinedEntries {
		sum = sum.add(DFP{e.Amount, e.AmountExp})
		fmt.Println("sum=", sum)

	}

	/*
		viewDistributionJoined : Language -> Int -> DFP -> DistributionJoined -> DRCRFormat -> Html Msg
		viewDistributionJoined language p runtot dj drcr =
		    tr []
		        ([ td [] [ text dj.tx_time ]
		         , td [] [ text dj.tx_notes ]
		         ]
		            ++ viewDFP (DFP dj.amount dj.amount_exp) p drcr
		            ++ viewDFP (DecimalFP.dfp_add (DFP dj.amount dj.amount_exp) runtot) p drcr
		            ++ [ td []
		                    [ a
		                        [ href ("/transactions/" ++ String.fromInt dj.tid)
		                        , class "button is-link"
		                        ]
		                        [ language |> tx_edit |> text ]
		                    ]
		               ]
		        )


		viewDistributionJoineds : Model.Model -> DFP -> List DistributionJoined -> List (Html Msg)
		viewDistributionJoineds model runtot distributionJoined =
		    case distributionJoined of
		        -- this function should not be called if there are no distributionJoineds, so this case should never happen.  But try tellin' that to Elm!
		        [] ->
		            [ tr [] [ td [] [ text "max fubar error" ] ] ]

		        [ x ] ->
		            [ viewDistributionJoined model.language model.accounts.decimalPlaces runtot x model.drcr_format ]

		        x :: xs ->
		            viewDistributionJoined model.language model.accounts.decimalPlaces runtot x model.drcr_format
		                :: viewDistributionJoineds model (DecimalFP.dfp_add (DFP x.amount x.amount_exp) runtot) xs
	*/

	for _, e := range config.CompareConfig.Funding {
		fmt.Println("e=", e)
		//for _, e2 := range e1.Currencies {
		//fmt.Println("e2=", e2)
		//}
	}

	// Get each balance in okex spot
	for _, e := range config.CompareConfig.Funding {
		fmt.Println("e=", e)
		//for _, e2 := range e1.Currencies {
		//fmt.Println("e2=", e2)
		//}
	}

	// 6. For each balance in an OKEx API result that is not requested by the config, we can create a Comparison object for okex only.

	// 7. For each balance in a Bookwerx API result that is not request by the config, we can create a Comparison object for bookwerx only.

	// In this way we can analyse the mass thus:

	// 1. Any Comparison that contains balances for OKEx _and_ Bookwerx are properly configured.  The balances may disagree, but that's a separate issue.  The essential issue is that the configure found them.

	// 2. Any Comparison that contains balances for OKEx _only_ are not properly configured.  The balances were seen by the OKEx API, but are not specified in the config.  This needs fixin'.

	// 3. Any Comparison that contains balances for Bookwerx _only_ are not properly configured either.  The balances were seen by the Bookwerx API, but are not specified in the config.  This needs fixin'.

	fmt.Println("balance in bookwerx....")
	// Get each balance in bookwerx

	// Is each okex balance present in bookwerx?

	// Is each bookwerx balance present in okex?

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
