// +build linux

package arp

import (
	"bufio"
	"net"
	"os"
	"strings"

	. "gitlab.x.lan/yunshan/droplet-libs/utils"
)

const (
	IP_ADDR = iota
	HW_TYPE
	FLAGS
	HW_ADDR
	MASK
	DEVICE
)

func GetTable() ArpTable {
	f, err := os.Open("/proc/net/arp")
	if err != nil {
		return nil
	}

	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip the field descriptions

	table := make(ArpTable)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		ip := net.ParseIP(fields[IP_ADDR]).To4()
		mac, _ := net.ParseMAC(fields[HW_ADDR])
		iface, _ := net.InterfaceByName(fields[DEVICE])
		table[IpToUint32(ip)] = ArpEntry{ip, mac, iface}
	}
	return table
}
