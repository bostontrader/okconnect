package config

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"io/ioutil"
)

// OKConnect needs to talk to an OKEx server and a bookwerx-core-rust server.
type Config struct {
	BookwerxConfig BookwerxConfig
	OKExConfig     OKExConfig
}

// What does OKConnect need to know in order to communicate with a bookwerx-core-rust server?
type BookwerxConfig struct {
	APIKey  string
	BaseURL string `yaml:"base_url"` // for example: http:185.183.96.73:3003

	// Any user account that is a...
	// ... funding account shall be tagged with this category
	CatFunding uint32 `yaml:"cat_funding"`

	// ... spot available account shall be tagged with this category
	CatSpotAvailable uint32 `yaml:"cat_spot_available"`

	// ... spot hold account shall be tagged with this category
	CatSpotHold uint32 `yaml:"cat_spot_hold"`

	// Any transaction that is a...
	// ... deposit into OKEx funding shall be tagged with this category
	CatDeposit uint32 `yaml:"cat_deposit"`
}

type OKExConfig struct {
	Credentials string
	BaseURL     string `yaml:"base_url"` // for example: https:www.okex.com
}

// Read the given credentials file for OKEx or the OKCatbox.
func ReadCredentialsFile(keyFile string) (*utils.Credentials, error) {
	var obj utils.Credentials
	data, err := ioutil.ReadFile(keyFile)
	if err != nil {
		fmt.Printf("compare.go:readCredentials: %v\n", err)
		return nil, err
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		fmt.Printf("compare.go:readCredentials: Cannot parse the credentials file.\n")
		return nil, err
	}
	return &obj, nil
}
