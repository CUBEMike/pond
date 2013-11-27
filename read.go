package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	response, _ := http.Get("http://localhost:12345")
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	var data []interface{}

	if err := json.Unmarshal(contents, &data); err != nil {
		panic(err)
	}

	for _, message := range data {
		fmt.Printf("%s", message)
	}
}
