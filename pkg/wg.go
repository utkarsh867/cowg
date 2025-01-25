package wg

import (
	"log"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func PrintDevices(devices []*wgtypes.Device) {
  for _, d := range devices {
    log.Printf("Device %s of type %s is listening on %d", d.Name, d.Type.String(), d.ListenPort)
  }
}

