// Package device provides access to hardware and system information.
package device

// TODO: Provide access to processor architecture.

import (
	"bytes"
	"fmt"
	snet "net"
	"time"

	"github.com/rdegges/go-ipify"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"

	"github.com/aukletio/Auklet-Client-C/errorlog"
)

// CurrentIP returns the device's current public IP address.
func CurrentIP() (ip string) {
	// Not covered in tests, because it depends on an external service.
	ip, err := ipify.GetIp()
	if err != nil {
		errorlog.Print(err)
	}
	return
}

// IfaceHash generates a unique device identifier based on the MAC addresses of
// hardware interfaces.
//
// I'm concerned that this isn't a good way to generate device identifiers.
// Alternatives: use a file to store a generated UUID.
func IfaceHash() string {
	// Not covered in tests, because it's unclear how to test for correctness.

	// MAC addresses are generally 6 bytes long
	sum := make([]byte, 6)
	interfaces, _ := snet.Interfaces()
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

// Metrics represents overall system metrics.
type Metrics struct {
	CPUPercent float64 `json:"cpuUsage"`
	MemPercent float64 `json:"memoryUsage"`
	rates
}

// rates contains the number of bytes per second of network traffic over all
// interfaces.
type rates struct {
	In  uint64 `json:"inboundNetwork"`
	Out uint64 `json:"outboundNetwork"`
}

func diff(cur, prev rates) rates {
	return rates{
		In:  cur.In - prev.In,
		Out: cur.Out - prev.Out,
	}
}

// serve generates a stream of network I/O rate values. The values sent on
// the stream are updated once per second; consuming at a higher rate than this
// will not increase resolution.
func (mon Monitor) serve() {
	defer close(mon.ioRates)
	var prev, cur rates
	update := func() {
		stats, err := net.IOCounters(false)
		if err == nil && len(stats) == 1 {
			stat := stats[0]
			prev = cur
			cur = rates{In: stat.BytesRecv, Out: stat.BytesSent}
		}
	}
	update()
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-mon.done:
			return
		case <-tick.C:
			update()
		case mon.ioRates <- diff(cur, prev):
		}
	}
}

// Monitor provides a source of Metrics.
type Monitor struct {
	ioRates chan rates
	done    chan struct{}
}

// NewMonitor returns a new Monitor.
func NewMonitor() Monitor {
	cpu.Percent(0, false)
	m := Monitor{
		ioRates: make(chan rates),
		done:    make(chan struct{}),
	}
	go m.serve()
	return m
}

// GetMetrics provides current system metrics.
func (mon Monitor) GetMetrics() Metrics {
	c, _ := cpu.Percent(0, false)
	m, _ := mem.VirtualMemory()
	return Metrics{
		rates:      <-mon.ioRates,
		CPUPercent: c[0],
		MemPercent: m.UsedPercent,
	}
}

// Close shuts down the Monitor.
func (mon Monitor) Close() { close(mon.done) }
