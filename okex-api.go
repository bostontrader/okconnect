// The purpose of this file is to hold items required to communicate with an OKEx/OKCatbox server.

package main

// Make the API call to transfer funds between OKEx accounts.
//func postTransfer(cfg Config, credentials utils.Credentials) (error) {

//urlBase := cfg.OKExConfig.Server
//endpoint := "/api/account/v3/wallet"
//url := urlBase + endpoint
//client := getHTTPClient(urlBase)
//timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
//prehash := timestamp + "POST" + endpoint
//encoded, _ := utils.HmacSha256Base64Signer(prehash, credentials.SecretKey)

//req, err := http.NewRequest("POST", url, nil)
//if err != nil {
//fmt.Println("NewRequest error: %v", err)
//return err
//}

//req.Header.Add("OK-ACCESS-KEY", credentials.Key)
//req.Header.Add("OK-ACCESS-SIGN", encoded)
//req.Header.Add("OK-ACCESS-TIMESTAMP", timestamp)
//req.Header.Add("OK-ACCESS-PASSPHRASE", credentials.Passphrase)
//resp, err := client.Do(req)
//if err != nil {
//fmt.Println("client.Do error: %v", err)
//return err
//}

//body, err := ioutil.ReadAll(resp.Body)
//if err != nil {
//fmt.Println("ReadAll error: %v", err)
//return err
//}

//_ = resp.Body.Close()

//if resp.StatusCode != 200 {
//fmt.Println("Status Code error: expected= 200, received=", resp.StatusCode)
//fmt.Println("body=", string(body))
//return errors.New("Status code error")
//}

//walletEntries := make([]utils.WalletEntry, 0)
//dec := json.NewDecoder(bytes.NewReader(body))
//dec.DisallowUnknownFields()
//err = dec.Decode(&walletEntries)
//if err != nil {
//fmt.Println("Wallet JSON decode error: %v", err)
//return nil, err
//}

//return walletEntries, nil
//}
