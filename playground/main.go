package main

import (
	"flag"
	"fmt"
	"os"
	"log"
	"io/ioutil"

	"github.com/tidwall/gjson"
)

func main() {
	data := flag.String("json", "data.json", "Data file to work with")
	req := flag.String("r", "", "JSON request")
	flag.Parse()

	jdata, err := os.Open(*data)
	if err != nil {
		log.Fatal(err)
	}
	defer jdata.Close()
	jbdata, err := ioutil.ReadAll(jdata)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "%s => %s\n", *req, gjson.GetBytes(jbdata, *req))
}

