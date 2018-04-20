// Package device provides access to hardware and system information.
package device

// TODO: Provide access to processor architecture.

import (
	"bytes"
	"fmt"
	"log"
	snet "net"
	"time"

	"github.com/rdegges/go-ipify"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

// CurrentIP returns the device's current public IP address.
func CurrentIP() (ip string) {
	ip, _ = ipify.GetIp()
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
		log.Print(err)
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

var inboundRate, outboundRate uint64

func init() {
	cpu.Percent(0, false)
	go networkStat()
}

// networkStat updates inbound and outbound network traffic figures
// periodically.
func networkStat() { // inboundRate outBoundRate
	// Total network I/O bytes recieved and sent per second from the system
	// since the start of the system.

	var inbound, outbound, inboundPrev, outboundPrev uint64
	for {
		if tempNet, err := net.IOCounters(false); err == nil {
			inbound = tempNet[0].BytesRecv
			outbound = tempNet[0].BytesSent
			inboundRate = inbound - inboundPrev
			outboundRate = outbound - outboundPrev
			inboundPrev = inbound
			outboundPrev = outbound
		}

		time.Sleep(time.Second)
	}
}

// Metrics represents overall system metrics.
type Metrics struct {
	CPUPercent float64 `json:"cpuUsage"`
	MemPercent float64 `json:"memUsage"`
	Inbound    uint64  `json:"inboundNetwork"`
	Outbound   uint64  `json:"outboundNetwork"`
}

// GetMetrics provides current system metrics.
func GetMetrics() (m Metrics) { // inboundRate outboundRate
	// System-wide cpu usage since the start of the child process
	if tempCPU, err := cpu.Percent(0, false); err == nil {
		m.CPUPercent = tempCPU[0]
	}

	// System-wide current virtual memory (ram) consumption
	// percentage.
	if tempMem, err := mem.VirtualMemory(); err == nil {
		m.MemPercent = tempMem.UsedPercent
	}

	m.Inbound = inboundRate
	m.Outbound = outboundRate
	return
}
