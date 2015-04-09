package main

import (
	"encoding/binary"
	"log"
	"os"

	"github.com/ligustah/dota2matchinfo/steam"

	"github.com/bitly/go-nsq"
	_ "github.com/codegangsta/envy/autoload"
	"github.com/syndtr/goleveldb/leveldb"

	"gopkg.in/vmihailenco/msgpack.v2"
)

func main() {
	db, err := leveldb.OpenFile(os.Getenv("LEVELDB_PATH"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	cfg := nsq.NewConfig()
	cfg.AuthSecret = os.Getenv("NSQ_SECRET")

	consumer, _ := nsq.NewConsumer("matches", "storage", cfg)
	consumer.AddHandler(nsq.HandlerFunc(func(msg *nsq.Message) error {

		match := new(steam.Match)

		err := msgpack.Unmarshal(msg.Body, match)
		if err != nil {
			return err
		}

		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, match.MatchID)

		return db.Put(key, msg.Body, nil)

	}))

	consumer.ConnectToNSQD(os.Getenv("NSQ_ADDRESS"))
	<-make(chan bool)
}
