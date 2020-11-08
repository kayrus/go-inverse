package inverse

import (
	"encoding/binary"
	"net"
	"sort"

	"github.com/EvilSuperstars/go-cidrman"
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

// SortCIDRs sorts a list of CIDRs.
func SortCIDRs(cidrs []*net.IPNet) {
	sort.SliceStable(cidrs, func(i, j int) bool {
		return ipv4ToUint32(cidrs[i].IP) < ipv4ToUint32(cidrs[j].IP)
	})
}

func findAcs(cidr *net.IPNet) []*net.IPNet {
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

func findDesc(cidr *net.IPNet) []*net.IPNet {
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

// inverseCIDR returns a list of CIDRs, which exclude the input CIDR.
func inverseCIDR(cidr *net.IPNet) []*net.IPNet {
	return append(findAcs(cidr), findDesc(cidr)...)
}

// InverseCIDR returns a sorted list of CIDRs, which exclude the input CIDR.
func InverseCIDR(cidr *net.IPNet) []*net.IPNet {
	cidrs := inverseCIDR(cidr)

	SortCIDRs(cidrs)

	return cidrs
}

// InverseCIDRs returns a sorted list of CIDRs, which exclude the input list of
// CIDRs.
func InverseCIDRs(cidrs []*net.IPNet) []*net.IPNet {
	var t []*net.IPNet
	for _, cidr := range cidrs {
		t = append(t, inverseCIDR(cidr)...)
	}

	res, _ := cidrman.MergeIPNets(t)

	return res
}
