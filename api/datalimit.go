package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/ESG-USA/Auklet-Client-C/config"
)

// CellularConfig defines a limit and date for devices that use a cellular
// connection.
type CellularConfig struct {
	// LimitPtr is a pointer to the maximum number of application layer
	// megabytes/period that the client may send over a cellular connection. If
	// nil, there is no limit. This field is provided only for serialization.
	// Clients should use Limit and LimitDefined instead.
	LimitPtr *int `json:"cellular_data_limit"`

	Limit int
	LimitDefined bool

	// Date is the day of the month that delimits a cellular
	// data plan period. Valid values are within [1, 28].
	Date int `json:"normalized_cell_plan_date"`
}

// DataLimit represents parameters that control the client's use of data.
type DataLimit struct {
	// EmissionPeriod is the time in seconds the client is to wait
	// between emission requests to the agent.
	EmissionPeriod int `json:"emission_period"`
	Storage        struct {
		// Limit is the maximum number of megabytes the client
		// may use to store unsent messages. If nil, there is no
		// storage limit.
		Limit *int `json:"storage_limit"`
	} `json:"storage"`
	Cellular CellularConfig `json:"data"`
}

// GetDataLimit returns a DataLimit from the dataLimit endpoint.
func GetDataLimit() (*DataLimit, error) {
	resp, err := get(fmt.Sprintf(dataLimitEP, config.AppID()), "application/json")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errStatus{resp}
	}

	var l struct {
		DataLimit `json:"config"`
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &l); err != nil {
		return nil, errEncoding{err, string(body), "GetDataLimit"}
	}

	depointerize(&l.DataLimit.Cellular)
	return &l.DataLimit, nil
}

func depointerize(c *CellularConfig) {
	if c.LimitPtr == nil {
		return
	}
	c.LimitDefined = true
	c.Limit = *c.LimitPtr
}
