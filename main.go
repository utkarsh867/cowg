package main

import (
	wg "github.com/utkarsh867/cowg/pkg"
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

	wg.PrintDevices(devices)

	log.Printf("Devices %d", len(devices))
}
