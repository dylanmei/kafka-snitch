package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
)

func main() {
	config := &SnitchConfig{}
	config.Parse()

	influxdbWriter, err := NewInfluxDBWriter(&config.InfluxDB)
	if err != nil {
		log.Panicf("Problem with InfluxDB config! %v", err)
	}

	observer := &Observer{
		writer: influxdbWriter,
	}

	snitch := NewSnitch(observer, &config.Observe)
	brokers := strings.Split(config.Brokers, ",")
	log.Infof("Starting snitch.")

	select {
	case <-snitch.Connect(brokers):
		go snitch.Run()
		break

	case <-time.After(60 * time.Second):
		log.Fatal("Couldn't start snitch! Quitting.")
	}

	termCh := make(chan os.Signal)
	signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)
	<-termCh

	log.Infof("Stopping snitch.")
	snitch.Close()
	observer.Close()

	log.Infof("Done.")
}
