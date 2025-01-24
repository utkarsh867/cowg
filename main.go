package main

import (
	"log"

	"golang.zx2c4.com/wireguard/wgctrl"
)

func main() {
  client, err := wgctrl.New()
  if err != nil {
    log.Fatal(err)
  }

  devices, err := client.Devices()
  if err != nil {
    log.Fatal(err)
  }

  log.Printf("Devices %d", len(devices))
}

