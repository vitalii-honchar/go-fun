package main

import (
	"github.com/charmbracelet/log"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{})
}