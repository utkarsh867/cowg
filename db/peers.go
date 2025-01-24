package db

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
  Client *gorm.DB
}

func Connect() (*DB, error) {
  db, err := gorm.Open(sqlite.Open("cowg.db"), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Silent),
  })
  if err != nil {
    return nil, err
  }

  db.AutoMigrate(&Peer{})

  return &DB {
    Client: db,
  }, nil
}

func (db *DB) AddPeer(p *wgtypes.PeerConfig, peerName string, privateKey string) {
  db.Client.Create(&Peer{
    Name: peerName,
    PublicKey: p.PublicKey.String(),
    PresharedKey: p.PresharedKey.String(),
    PrivateKey: privateKey,
  })
}

func (db *DB) DeletePeer(p *wgtypes.PeerConfig) {
  db.Client.Where(&Peer{
    PublicKey: p.PublicKey.String(),
  }).Delete(&Peer{})
}

func (db *DB) GetPeer(p *wgtypes.PeerConfig) *Peer {
  var peer Peer
  db.Client.Where(&Peer{
    PublicKey: p.PublicKey.String(),
  }).Find(&peer)
  return &peer
}
