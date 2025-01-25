package cowg

import (
	"log"
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

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

  log.Printf("Private Key: %s", privateKey.String())
  return nil
}

func DeletePeer(c *Cowg, peer string) error {
  publicKey, err := wgtypes.ParseKey(peer)
  if err != nil {
    return err
  }

  c.WgClient.ConfigureDevice(c.WgDevice.Name, wgtypes.Config{
    Peers: []wgtypes.PeerConfig{
      wgtypes.PeerConfig {
        PublicKey: publicKey,
        Remove: true,
      },
    },
  })
  return nil 
}
