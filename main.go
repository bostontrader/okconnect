package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
)

type BookwerxConfig struct {
	APIKey string
	Server string
}

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
	help := flag.Bool("help", false, "Display a help screen")
	config := flag.String("config", "/path/to/config.yml", "The config file for OKConnect")

	// Args[0] is the path to the program
	// Args[1] is okconnect
	// Args[2:] are any remaining args.
	if len(os.Args) < 2 { // Invoke w/o any args
		flag.Usage()
	} else {

		flag.Parse()
		fmt.Println("OKConnect is using the following runtime args:")
		fmt.Println("cmd:", *cmd)
		fmt.Println("help:", *help)
		fmt.Println("config:", *config)

		// Try to read the config file.
		data, err := ioutil.ReadFile(*config)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		t := Config{}

		err = yaml.Unmarshal([]byte(data), &t)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		switch *cmd {
		case "compare":
			Compare(t)

		case "help":
			fmt.Println("Available commands:")
			fmt.Println("help\tGuess what this command does?")
			fmt.Println("compare\tCompare the balances between OKEx and Bookwerx")

		default:
			fmt.Println("Unknown command ", *cmd)
		}
	}

}
