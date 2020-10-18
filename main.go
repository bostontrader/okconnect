package main

import (
	"flag"
	"fmt"
	"github.com/bostontrader/okconnect/compare"
	"github.com/bostontrader/okconnect/config"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("")
	fmt.Println("    okconnect <command> [arguments]")
	fmt.Println("")
	fmt.Println("The commands are:")
	fmt.Println("    compare, transfer")
	fmt.Println("")
	fmt.Println("Use \"okconnect <command>\" without any arguments to see more info about that command.")
}

func readConfigFile(filename *string) (cfg *config.Config, err error) {
	data, err := ioutil.ReadFile(*filename)
	if err != nil {
		fmt.Printf("ReadFile error: %v\n", err)
		return nil, err
	}

	cfg = &config.Config{}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		fmt.Printf("Cannot parse config file.\n")
		return nil, err
	}

	return
}

func main() {

	compareCmd := flag.NewFlagSet("compare", flag.ExitOnError)
	compareConfig := compareCmd.String("config", "/path/to/config.yml", "The config file for OKConnect")

	// okconnect transfer -currency BTC -quan 1.25 -from 6 -to 3 -config okconnect.yaml
	transferCmd := flag.NewFlagSet("transfer", flag.ExitOnError)
	transferConfig := transferCmd.String("config", "/path/to/config.yml", "The config file for OKConnect")
	transferCurrency := transferCmd.String("currency", "BTC", "Which currency to transfer")
	transferQuan := transferCmd.String("quan", "0.0", "How much to transfer")
	transferFrom := transferCmd.String("from", "3", "Source: \"1\" (spot) or \"6\" (funding)")
	transferTo := transferCmd.String("to", "3", "Destination: \"1\" (spot) or \"6\" (funding)")

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
				err := compareCmd.Parse(os.Args[2:])
				if err != nil {
					fmt.Printf("Cannot parse the args.\n")
					return
				}

				cfg, err := readConfigFile(compareConfig)
				if err != nil {
					fmt.Printf("Cannot read the config file.\n")
					return
				}
				compare.Compare(cfg)
			}

		case "transfer":
			if len(os.Args) <= 2 { // Invoked with this command but w/o any other args
				transferCmd.Usage()
			} else {
				err := transferCmd.Parse(os.Args[2:])
				if err != nil {
					fmt.Printf("Cannot parse the args.\n")
					return
				}

				cfg, err := readConfigFile(transferConfig)
				if err != nil {
					fmt.Printf("Cannot read the config file.\n")
					return
				}
				Transfer(cfg, transferCurrency, transferFrom, transferTo, transferQuan)
			}

		default:
			fmt.Printf("The command %s is not defined.\n", os.Args[1])
			printUsage()
		}

	}
}
