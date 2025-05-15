package main

import (
	"context"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	connectionTimeout = 5 * time.Second
	readTimeout       = 5 * time.Second
)

func scanAddress(address string) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		conn, err := net.DialTimeout("tcp", address, connectionTimeout)
		if err != nil {
			log.WithField("address", address).
				WithError(err).
				Error("Error connecting to address")
			return
		}
		defer conn.Close()

		log.WithField("address", address).Info("Connection successful")

		ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
		defer cancel()

		var res string
		buf := make([]byte, 1024)

		for {
			select {
			case <-ctx.Done():
				log.WithField("address", address).Info("Read timeout")
				ch <- res
				return
			default:
				conn.SetReadDeadline(time.Now().Add(1 * time.Second))

				n, err := conn.Read(buf)
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						log.Info("No data received (timeout)")
						continue
					}
					log.Error("Read error")
					ch <- res
					return
				}

				res += string(buf[:n])
			}
		}
	}()

	return ch
}

func main() {
	log.SetFormatter(&log.TextFormatter{})

	log.Info("Starting port scanner")

	conn, err := net.DialTimeout("tcp", "127.0.0.1:8080", connectionTimeout)
	if err != nil {
		log.Error("Error connecting to port: ", err)
		return
	}
	defer conn.Close()

	log.Info("Connection successful")

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	buf := make([]byte, 1024)

	for {
		select {
		case <-ctx.Done():
			log.Info("Read timeout")
			return
		default:
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			n, err := conn.Read(buf)
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					log.Info("No data received (timeout)")
					continue
				}
				log.Error("Read error")
				return
			}

			log.Printf("Read: %d bytes: %s", n, buf[:n])
		}
	}
}
