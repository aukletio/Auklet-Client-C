package api

import (
	"encoding/json"
	"io"
	"log"
)

// DataLimit represents parameters that control the client's use of data.
type DataLimit struct {
	// Cellular is the maximum number of application layer megabytes/period
	// that the client may send over a cellular connection.
	// PlanDate is the day of the month that delimits a cellular data plan
	// period. Valid values are within [1, 28]. If either field is nil,
	// there is no limit.
	Cellular *int `json:"cellular_data_limit"`
	PlanDate *int `json:"plan_date"`

	// Storage is the maximum number of gigabytes the client may use to
	// store unsent messages. If nil, there is no storage limit.
	Storage *int `json:"storage_limit"`

	// EmissionPeriod is the time in seconds the client is to wait between
	// emission requests to the agent. If nil, the client self-destructs.
	EmissionPeriod *int `json:"emission_period"`
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
