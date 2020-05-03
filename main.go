package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	A string
	B struct {
		RenamedC int   `yaml:"c"`
		D        []int `yaml:",flow"`
	}
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

	data, err := ioutil.ReadFile(*config)
	if err != nil {
		fmt.Print(err)
	}

	t := Config{}

	err = yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", t)

	d, err := yaml.Marshal(&t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t dump:\n%s\n\n", string(d))

	m := make(map[interface{}]interface{})

	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m:\n%v\n\n", m)

	d, err = yaml.Marshal(&m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m dump:\n%s\n\n", string(d))

	switch *cmd {
	case "help":
		fmt.Println("Available commands:")
		fmt.Println("help\tGuess what this command does?")
		fmt.Println("compare\tCompare the balances between OKEx and Bookwerx")

	case "compare":
		fmt.Println("balance in okex....")
		fmt.Println("balance in bookwerx....")

	default:
		fmt.Println("Unknown endpoint ", *cmd)
	}
}
