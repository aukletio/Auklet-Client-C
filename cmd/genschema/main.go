// Command genschema prints an example of each producer schema to stdout.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/ESG-USA/Auklet-Client-C/schema"
)

func errorsig() {
	agentdata := `[
		{
			"functionAddress": 0,
			"callSiteAddress": 0
		},
		{
			"functionAddress": 0,
			"callSiteAddress": 0
		}
	]`
	b, err := json.MarshalIndent(schema.ErrorSig{
		Trace: json.RawMessage(agentdata),
	}, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println("ErrorSig:")
	fmt.Println(string(b))
}

func exit() {
	fmt.Println("Exit:")
	b, err := json.MarshalIndent(schema.Exit{}, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func profile() {
	treedata := `{
		"functionAddress": 0,
		"callSiteAddress": 0,
		"nCalls":1,
		"nSamples":1,
		"callees":[
			{
				"functionAddress": 0,
				"callSiteAddress": 0,
				"nCalls":1,
				"nSamples":1,
				"callees":[]
			},
			{
				"functionAddress": 0,
				"callSiteAddress": 0,
				"nCalls":1,
				"nSamples":1,
				"callees":[]
			}
		]
	}`
	fmt.Println("Profile:")
	b, err := json.MarshalIndent(schema.Profile{
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
