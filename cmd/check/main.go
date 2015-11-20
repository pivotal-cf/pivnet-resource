package main

import (
	"encoding/json"
	"os"
)

type input struct {
	Source struct {
		APIToken     string `json:"api_token"`
		ResourceName string `json:"resource_name"`
	} `json:"source"`
}

type version []map[string]string

func main() {
	var i input

	err := json.NewDecoder(os.Stdin).Decode(&i)
	if err != nil {
		panic(err)
	}

	v := version{}
	switch i.Source.ResourceName {
	case "p-gitlab":
		v = append(v, map[string]string{"version": "0.1.1 BETA"})
	default:
		v = append(v, map[string]string{"version": "1.7.1.0"})
	}

	err = json.NewEncoder(os.Stdout).Encode(v)
	if err != nil {
		panic(err)
	}
}
