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

	log.Info("Starting snitch")
	snitch := NewSnitch(observer, &config.Observe)
	brokers := strings.Split(config.Brokers, ",")

	select {
	case <-snitch.Connect(brokers):
		break

	case <-time.After(60 * time.Second):
		log.Fatal("Couldn't start snitch! Quitting")
	}

	if config.RunOnce {
		snitch.RunOnce()
	} else {
		go snitch.Run(config.RunSnooze)

		termCh := make(chan os.Signal)
		signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)
		<-termCh

		log.Info("Stopping snitch")
		snitch.Close()
	}

	observer.Flush()
	log.Info("Done!")
}
