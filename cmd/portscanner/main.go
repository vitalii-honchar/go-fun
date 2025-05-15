package main

import (
	"context"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	connectionTimeout = 5 * time.Second
	readTimeout       = 5 * time.Second
	portStart         = 1
	portEnd           = 9000
)

type ScanResult struct {
	Address string
	Result  string
}

func scanAddress(address string) <-chan *ScanResult {
	ch := make(chan *ScanResult)

	go func() {
		defer close(ch)

		conn, err := net.DialTimeout("tcp", address, connectionTimeout)
		if err != nil {
			// log.WithField("address", address).
			// 	WithError(err).
			// 	Error("Error connecting to address")
			return
		}
		defer conn.Close()

		log.WithField("address", address).Info("Connection successful")

		ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
		defer cancel()

		res := &ScanResult{Address: address, Result: ""}
		buf := make([]byte, 1024)

		for {
			select {
			case <-ctx.Done():
				// log.WithField("address", address).Info("Read timeout")
				ch <- res
				return
			default:
				conn.SetReadDeadline(time.Now().Add(1 * time.Second))

				n, err := conn.Read(buf)
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						// log.Info("No data received (timeout)")
						continue
					}
					// log.Error("Read error")
					ch <- res
					return
				}

				res.Result += string(buf[:n])
			}
		}
	}()

	return ch
}

func generateIPs(cidr string) ([]string, error) {
	// Parse the CIDR notation
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	// Convert IP to 4-byte representation
	ip = ip.To4()
	if ip == nil {
		return nil, fmt.Errorf("not an IPv4 address")
	}

	// Get the network and broadcast addresses
	mask := ipnet.Mask
	network := ip.Mask(mask)
	broadcast := make(net.IP, len(network))
	for i := 0; i < len(network); i++ {
		broadcast[i] = network[i] | ^mask[i]
	}

	// Increment IP until we reach broadcast address
	for ip := network; !ip.Equal(broadcast); inc(ip) {
		ips = append(ips, ip.String())
	}

	return ips, nil
}

// Helper function to increment IP address
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func generateScanAddresses(ips []string) []string {
	var addresses []string
	for _, ip := range ips {
		for port := portStart; port <= portEnd; port++ {
			addresses = append(addresses, fmt.Sprintf("%s:%d", ip, port))
		}
	}

	return addresses
}

func main() {
	log.SetFormatter(&log.TextFormatter{})

	log.Info("Starting port scanner")

	cidr := "192.168.1.0/24"
	ips, err := generateIPs(cidr)
	if err != nil {
		log.Fatal("unexpected error", err)
	}

	addresses := generateScanAddresses(ips)

	log.WithField("addresses", len(addresses)).Info("Scanning addresses")

	var results []<-chan *ScanResult
	for _, address := range addresses {
		results = append(results, scanAddress(address))
	}

	for _, resChannel := range results {
		res, ok := <-resChannel
		if ok {
			log.WithField("address", res.Address).
				WithField("result", res.Result).
				Info("Scanned address")
		}
	}
}
