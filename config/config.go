package config

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"io/ioutil"
)

type Config struct {
	BookwerxConfig BookwerxConfig
	OKExConfig     OKExConfig
}

// This is the configuration for a bookwerx core server and apikey for an ordinary user.
type BookwerxConfig struct {
	APIKey string
	Server string

	// Any user account that is a...
	// ... funding account shall be tagged with this category
	FundingCat int32 `yaml:"funding_cat"`

	// ... spot available account shall be tagged with this category
	SpotAvailableCat int32 `yaml:"spot_available_cat"`

	// ... spot hold account shall be tagged with this category
	SpotHoldCat int32 `yaml:"spot_hold_cat"`
}

type OKExConfig struct {
	Credentials string
	Server      string
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
