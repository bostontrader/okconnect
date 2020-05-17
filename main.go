package main

import (
	"flag"
	"fmt"
	//"log"
	"os"
)

// What account type.  Funding, Spot.
type Category struct {
	Category   string
	Currencies []Cur
}

type Cur struct {
	CurrencyID string
	Hold       int
	Available  int
}

type BookwerxConfig struct {
	APIKey string
	Server string
}

type OKExConfig struct {
	Credentials string
	Server      string
}

type CompareConfig struct {
	Funding []Cur
	Spot    []Cur
}
type Config struct {
	BookwerxConfig BookwerxConfig
	OKExConfig     OKExConfig
	CompareConfig  CompareConfig
}

func main() {

	cmd := flag.String("cmd", "help", "An OKConnect command")
	config := flag.String("config", "/path/to/config.yml", "The config file for OKConnect")
	okCredentialsFile := flag.String("okCredentialsFile", "/path/to/ok_credentials.json", "A file that contains the OKEx API credentials")

	if len(os.Args) < 2 {
		flag.Usage()
	}

	flag.Parse()
	fmt.Println("cmd:", *cmd)
	fmt.Println("config:", *config)
	fmt.Println("keyfile:", *okCredentialsFile)

	switch *cmd {
	case "help":
		fmt.Println("Available commands:")
		fmt.Println("help\tGuess what this command does?")
		fmt.Println("compare\tCompare the balances between OKEx and Bookwerx")

	default:
		fmt.Println("Unknown command ", *cmd)
	}
}
