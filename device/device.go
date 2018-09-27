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

	"github.com/ESG-USA/Auklet-Client-C/errorlog"
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

var ioRates = monitorRates()

func init() {
	cpu.Percent(0, false)
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

// monitorRates returns a stream of network I/O rate values. The values sent on
// the stream are updated once per second; consuming at a higher rate than this
// will not increase resolution.
func monitorRates() <-chan rates {
	r := make(chan rates)
	go func() {
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
		rates:      <-ioRates,
		CPUPercent: c[0],
		MemPercent: m.UsedPercent,
	}
}
