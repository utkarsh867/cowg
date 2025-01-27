package cowg

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/ini.v1"
)

func CreateWgClient() *wgctrl.Client {
  client, err := wgctrl.New()
  if err != nil {
    log.Fatal(err)
  }
  return client
}

func CreatePeer(c *Cowg, name string, address net.IP) (error) {
  privateKey, err := wgtypes.GeneratePrivateKey()
  if err != nil {
    return err
  }
  publicKey := privateKey.PublicKey()
  presharedKey, err := wgtypes.GenerateKey()
  if err != nil {
    return err
  }

  peer := wgtypes.PeerConfig{
    PublicKey: publicKey,
    PresharedKey: &presharedKey,
    AllowedIPs: []net.IPNet{
      net.IPNet{
        IP: address,
        Mask: net.IPv4Mask(255,255,255,255),
      },
    },
    Remove: false,
  }
  c.WgClient.ConfigureDevice(c.WgDevice.Name, wgtypes.Config{
    ReplacePeers: false,
    Peers: []wgtypes.PeerConfig{
      peer,
    },
  })
  c.Db.AddPeer(&peer, name, privateKey.String())
  return nil
}

func DeletePeer(c *Cowg, peer string) error {
  publicKey, err := wgtypes.ParseKey(peer)
  if err != nil {
    return err
  }

  peerConfig := wgtypes.PeerConfig {
    PublicKey: publicKey,
    Remove: true,
  }
  c.WgClient.ConfigureDevice(c.WgDevice.Name, wgtypes.Config{
    Peers: []wgtypes.PeerConfig{
      peerConfig,
    },
  })

  c.Db.DeletePeer(&peerConfig)
  return nil 
}

func PeerConfig(c *Cowg, p PeerListItem) (string, error) {
  publicKey, err := wgtypes.ParseKey(p.PublicKey())
  if err != nil {
    return "", err
  }
  peer := c.Db.GetPeer(&wgtypes.PeerConfig{
    PublicKey: publicKey,
  })
  cfg := ini.Empty()

  // [Interface]
  cfg.NewSection("Interface")
  cfg.Section("Interface").NewKey("PrivateKey", peer.PrivateKey)
  cfg.Section("Interface").NewKey("Address", p.Description())

  // [Peer]
  cfg.NewSection("Peer")
  cfg.Section("Peer").NewKey("PublicKey", c.WgDevice.PublicKey.String())
  cfg.Section("Peer").NewKey("PresharedKey", peer.PresharedKey)
  cfg.Section("Peer").NewKey("AllowedIPs", "0.0.0.0/0")
  cfg.Section("Peer").NewKey("Endpoint", fmt.Sprintf("10.8.0.1:%s", strconv.Itoa(c.WgDevice.ListenPort)))
  
  var b bytes.Buffer
  cfg.WriteTo(&b)

  return b.String(), nil
}
