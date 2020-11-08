package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"sort"
)

const (
	ipMin uint32 = 0
	ipMax uint32 = ^ipMin
)

func ipv4ToUint32(ip []byte) uint32 {
	return binary.BigEndian.Uint32(net.IP(ip).To4())
}

func uint32ToIPv4(i uint32) net.IP {
	ip := make(net.IP, net.IPv4len)
	binary.BigEndian.PutUint32(ip, i)
	return ip
}

func lastAddress(ipNet *net.IPNet) uint32 {
	return ipv4ToUint32(ipNet.IP) | ^ipv4ToUint32(ipNet.Mask)
}

func sortCIDRs(cidrs []*net.IPNet) {
	sort.SliceStable(cidrs, func(i, j int) bool {
		return ipv4ToUint32(cidrs[i].IP) < ipv4ToUint32(cidrs[j].IP)
	})
}

func findPrev(cidr *net.IPNet) []*net.IPNet {
	var cidrs []*net.IPNet
	firstIP := ipv4ToUint32(cidr.IP)
	addr := ipMin
	hostmask := ipMax
	for addr != firstIP {
		if v := addr | hostmask; v < firstIP {
			cidr := &net.IPNet{
				IP:   uint32ToIPv4(addr),
				Mask: net.IPMask(uint32ToIPv4(^hostmask)),
			}
			cidrs = append(cidrs, cidr)
			addr = v + 1
		}
		hostmask = hostmask >> 1
	}
	return cidrs
}

func findNext(cidr *net.IPNet) []*net.IPNet {
	var cidrs []*net.IPNet
	lastIP := lastAddress(cidr)
	addr := ipMax
	hostmask := ipMax
	for addr != lastIP {
		if v := addr ^ hostmask; v > lastIP {
			cidr := &net.IPNet{
				IP:   uint32ToIPv4(v),
				Mask: net.IPMask(uint32ToIPv4(^hostmask)),
			}
			cidrs = append(cidrs, cidr)
			addr = v - 1
		}
		hostmask = hostmask >> 1
	}
	return cidrs
}

func ReverseCIDRs(cidr *net.IPNet) []*net.IPNet {
	cidrs := append(findPrev(cidr), findNext(cidr)...)

	sortCIDRs(cidrs)

	return cidrs
}

func main() {
	flag.Parse()
	arg := flag.Arg(0)
	if arg == "" {
		log.Fatalf("CIDR argument is required")
	}
	ip, ipNet, err := net.ParseCIDR(arg)
	if err != nil {
		log.Fatal(err)
	}

	firstIP := ipv4ToUint32(ipNet.IP)
	lastIP := lastAddress(ipNet)

	cidrs := ReverseCIDRs(ipNet)

	for _, v := range cidrs {
		fmt.Println(v)
	}

	log.Printf("IP: %s, NET: %+#v", ip, ipNet)
	log.Printf("IPfirst: %s, IPlast: %s", uint32ToIPv4(firstIP), uint32ToIPv4(lastIP))
	return
}
