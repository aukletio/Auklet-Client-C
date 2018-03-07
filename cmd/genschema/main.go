// Command genschema prints an example of each client JSON schema to stdout.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/ESG-USA/Auklet-Client/schema"
)

func errorsig() {
	libaukletjson := `[
		{
			"fn": 0,
			"cs": 0
		},
		{
			"fn": 0,
			"cs": 0
		}
	]`
	b, err := json.MarshalIndent(schema.ErrorSig{
		Trace: json.RawMessage(libaukletjson),
	}, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println("ErrorSig:")
	fmt.Println(string(b))
}

func exit() {
	fmt.Println("Exit:")
	b, err = json.MarshalIndent(schema.Exit{}, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func profile() {
	treedata := `{
		"fn":0,
		"cs":0,
		"ncalls":1,
		"nsamples":1,
		"callees":[
			{
				"fn":0,
				"cs":0,
				"ncalls":1,
				"nsamples":1,
				"callees":[]
			},
			{
				"fn":0,
				"cs":0,
				"ncalls":1,
				"nsamples":1,
				"callees":[]
			}
		]
	}`
	fmt.Println("Profile:")
	b, err = json.MarshalIndent(schema.Profile{
		Tree: json.RawMessage(treedata),
	}, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func main() {
	errorsig()
	exit()
	profile()
}
