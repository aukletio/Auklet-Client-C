// Package device provides access to hardware and system information.
package device

// TODO: Provide access to processor architecture.

import (
	"bytes"
	"fmt"
	snet "net"
	"os"
	"time"

	"github.com/rdegges/go-ipify"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"

	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// CurrentIP returns the device's current public IP address.
func CurrentIP() (ip string) {
	ip, err := ipify.GetIp()
	if err != nil {
		errorlog.Print(err)
	}
	return
}

// MacHash is derived from the MAC addresses of all available network
// interfaces. It serves as a unique device identifier.
var MacHash = ifacehash()

// ifacehash generates a unique device identifier based on the MAC addresses of
// hardware interfaces.
//
// I'm concerned that this isn't a good way to generate device identifiers.
// Alternatives: use a file to store a generated UUID.
func ifacehash() string {
	// MAC addresses are generally 6 bytes long
	sum := make([]byte, 6)
	interfaces, err := snet.Interfaces()
	if err != nil {
		errorlog.Print(err)
	}

	for _, i := range interfaces {
		if bytes.Compare(i.HardwareAddr, nil) == 0 {
			continue
		}
		for h, k := range i.HardwareAddr {
			sum[h] += k
		}
	}
	return fmt.Sprintf("%x", string(sum))
}

var rates = monitorRates()

var inboundRate, outboundRate uint64

func init() {
	cpu.Percent(0, false)
}

// Metrics represents overall system metrics.
type Metrics struct {
	CPUPercent float64 `json:"cpuUsage"`
	MemPercent float64 `json:"memoryUsage"`
	Rates
}

// Rates contains the number of bytes per second of network traffic over all
// interfaces.
type Rates struct {
	In  uint64 `json:"inboundNetwork"`
	Out uint64 `json:"outboundNetwork"`
}

func diff(cur, prev Rates) Rates {
	return Rates{
		In:  cur.In - prev.In,
		Out: cur.Out - prev.Out,
	}
}

// monitorRates returns a stream of network I/O rate values. The values sent on
// the stream are updated once per second; consuming at a higher rate than this
// will not increase resolution.
func monitorRates() <-chan Rates {
	r := make(chan Rates)
	go func() {
		var prev, cur Rates
		update := func() {
			stats, err := net.IOCounters(false)
			if err != nil || len(stats) != 1 {
				// something isn't right
				return
			}
			stat := stats[0]
			prev = cur
			cur = Rates{In: stat.BytesRecv, Out: stat.BytesSent}
		}
		tick := time.Tick(time.Second)
		for {
			select {
			case <-tick:
				update()
			case r <- diff(cur, prev):
			}
		}
	}()
	return r
}

// GetMetrics provides current system metrics.
func GetMetrics() Metrics {
	c, _ := cpu.Percent(0, false)
	m, _ := mem.VirtualMemory()
	return Metrics{
		Rates:      <-rates,
		CPUPercent: c[0],
		MemPercent: m.UsedPercent,
	}
}
