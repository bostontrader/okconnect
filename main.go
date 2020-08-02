package main

import (
	"flag"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"

	//"flag"
	"fmt"
	"os"
)

// This is the configuration for a bookwerx core server and apikey for an ordinary user.
type BookwerxConfig struct {
	APIKey string
	Server string

	// Any user account that is a funding account shall be tagged with this category
	FundingCat int32 `yaml:"funding_cat"`
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

//type CompareConfig struct {
//Funding []Cur
//Spot    []Cur
//}
type Config struct {
	BookwerxConfig BookwerxConfig
	OKExConfig     OKExConfig
	//CompareConfig  CompareConfig
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("")
	fmt.Println("    okconnect <command> [arguments]")
	fmt.Println("")
	fmt.Println("The commands are:")
	fmt.Println("    compare")
	fmt.Println("")
	fmt.Println("Use \"okconnect <command>\" without any arguments to see more info about that command.")
}

func readConfigFile(filename *string) (cfg Config) {
	data, err := ioutil.ReadFile(*filename)
	if err != nil {
		log.Fatalf("error: %v", err)
		return
	}

	cfg = Config{}

	err = yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		log.Fatalf("error: %v", err)
		return
	}

	return
}

func main() {

	compareCmd := flag.NewFlagSet("compare", flag.ExitOnError)
	compareConfig := compareCmd.String("config", "/path/to/config.yml", "The config file for OKConnect")

	// Args[0] is okconnect
	// Args[1] should be a subcommand
	// Args[2:] are any remaining args.
	if len(os.Args) <= 1 { // Invoked w/o any args
		printUsage()
	} else {
		switch os.Args[1] { // this should be the subcommand
		case "compare":
			if len(os.Args) <= 2 { // Invoked with this command but w/o any other args
				compareCmd.Usage()
			} else {
				compareCmd.Parse(os.Args[2:])

				fmt.Printf("OKConnect is executing the %s command using the following runtime args:\n", os.Args[1])
				fmt.Println("config:", *compareConfig)
				Compare(readConfigFile(compareConfig))
			}
		default:
			fmt.Printf("The command %s is not defined.\n", os.Args[1])
			printUsage()
		}

	}
}
