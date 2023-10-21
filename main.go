package main

import (
	"context"
	"homeserver/config"
	"homeserver/home"
	"homeserver/webserver"
	logger "log"
	"net"
	"os/signal"
	"syscall"
)

var log *logger.Logger = logger.New(logger.Writer(), "[MAIN] ", logger.LstdFlags|logger.Lmsgprefix)

func init() {
	config.SetInt("port", 80)

	iface, err := net.InterfaceByName("en0")
	if err != nil {
		log.Panicf("could not get interface 'en0': %+v", err)
	}
	config.SetString("macAddr", iface.HardwareAddr.String())
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer stop()

	// go udp.Mcast()

	// webserver
	go func() {
		log.Print("Starting webserver...")
		err := webserver.Run()
		if err != nil {
			log.Fatalf("Could not start webserver: %+v", err)
		}
		log.Print("Stated the webserver!")
	}()

	home.AdvertiseSmartDevices()

	<-ctx.Done()
}
