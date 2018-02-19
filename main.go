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

	observer := NewObserver()
	if config.CanWriteToStatsD() {
		writer, err := NewStatsDWriter(&config.StatsD)
		if err != nil {
			log.Panicf("Problem with StatsD config! %v", err)
		}

		observer.AddWriter(writer)
	}

	if config.CanWriteToInfluxDB() {
		writer, err := NewInfluxDBWriter(&config.InfluxDB)
		if err != nil {
			log.Panicf("Problem with InfluxDB config! %v", err)
		}

		observer.AddWriter(writer)
	}

	snitch := NewSnitch(observer, &config.Observe)
	brokers := strings.Split(config.Brokers, ",")
	log.Infof("Starting snitch")

	select {
	case <-snitch.Connect(brokers):
		go snitch.Run()
		break

	case <-time.After(60 * time.Second):
		log.Fatal("Couldn't start snitch! Quitting")
	}

	termCh := make(chan os.Signal)
	signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)
	<-termCh

	log.Infof("Stopping snitch")
	snitch.Close()
	observer.Flush()

	log.Infof("Done!")
}
