package home

import (
	"fmt"
	"homeserver/config"
	logger "log"

	"github.com/koron/go-ssdp"
)

var log *logger.Logger = logger.New(logger.Writer(), "[Smart Device] ", logger.LstdFlags|logger.Lmsgprefix)
var ad *ssdp.Advertiser

func AdvertiseSmartDevices() {
	st := "urn:schemas-upnp-org:device:basic:1"
	usn := fmt.Sprintf("uuid:2f402f80-da50-11e1-9b23-%s::upnp:rootdevice", config.GetString("macAddr"))
	location := fmt.Sprintf("http://%s:%d/%s", config.GetString("ip"), config.GetInt("port"), "description.xml")
	server := "FreeRTOS/6.0.5, UPnP/1.0, IpBridge/1.17.0"

	var err error
	ad, err = ssdp.Advertise(st, usn, location, server, 1800)
	if err != nil {
		log.Printf("Could not avertise device: %+v", err)
		return
	}

	return
}

func CloseSmartDeviceAdvertiser() error {
	if ad == nil {
		return nil
	}
	if err := ad.Bye(); err != nil {
		return err
	}
	return ad.Close()
}
