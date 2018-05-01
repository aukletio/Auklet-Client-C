package api

import (
	"encoding/json"
	"io"
	"log"
)

// DataLimit represents parameters that control the client's use of data.
type DataLimit struct {
	// Cellular is the number of megabytes/period that the client may send
	// over a cellular connection.
	Cellular       int `json:"cellular_data_limit"` // MB/period
	Storage        int `json:"storage_limit"`       // GB
	EmissionPeriod int `json:"emission_rate"`       // seconds
	PlanDate       int `json:"plan_date"`           // day of the month ∈ [1, 28]
}

// DataLimit returns a DataLimit from the dataLimit endpoint.
func (api API) DataLimit() (l DataLimit) {
	resp := api.get(dataLimit, "application/json")
	if resp == nil {
		return
	}
	if resp.StatusCode != 200 {
		log.Printf("api.DataLimit: unexpected status %v", resp.Status)
	}
	err := json.NewDecoder(resp.Body).Decode(&l)
	if err != nil && err != io.EOF {
		log.Print(err)
	}
	return
}
