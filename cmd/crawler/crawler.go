package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ligustah/dota2matchinfo/steam"

	"github.com/bitly/go-nsq"
	"github.com/cenkalti/backoff"
	_ "github.com/codegangsta/envy/autoload"

	"gopkg.in/mgo.v2/bson"
	"gopkg.in/vmihailenco/msgpack.v2"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var (
	steam_api_key = os.Getenv("STEAM_API_KEY")
)

type Batch struct {
	Id bson.ObjectId `json:"-" bson:"_id,omitempty"`

	Start     uint64    `json:"start" bson:"start"`
	End       uint64    `json:"end" bson:"end"`
	Created   time.Time `json:"created" bson:"created"`
	Completed time.Time `json:"completed" bson:"completed"`
	Attempts  uint64    `json:"attempts" bson:"attempts"`
	Worker    string    `json:"worker" bson:"worker"`
	UserAgent string    `json:"useragent" bson:"useragent"`
}

func NewMatchFetcher(cfg MatchFetcherConfig) (mf *MatchFetcher, err error) {
	mf = new(MatchFetcher)
	mf.api = steam.NewApi(cfg.SteamApiKey)

	c := nsq.NewConfig()
	c.MaxInFlight = 1
	c.AuthSecret = os.Getenv("WORKER_SECRET")
	c.UserAgent = "go-crawler"
	c.MsgTimeout = 30 * time.Minute

	mf.consumer, err = nsq.NewConsumer("batches", "worker", c)
	if err != nil {
		return
	}

	mf.consumer.AddHandler(mf)
	err = mf.consumer.ConnectToNSQD(cfg.NSQDAddress)
	if err != nil {
		return
	}

	mf.producer, err = nsq.NewProducer(cfg.NSQDAddress, c)

	return
}

type MatchFetcherConfig struct {
	LocalAddress string
	SteamApiKey  string
	NSQDAddress  string
}

type MatchFetcher struct {
	api      steam.Api
	consumer *nsq.Consumer
	producer *nsq.Producer
}

func (mf *MatchFetcher) HandleMessage(msg *nsq.Message) error {
	//each message is a batch for us to process

	var (
		resp       *http.Response
		req        *http.Request
		client     *http.Client
		expBackoff *backoff.ExponentialBackOff = backoff.NewExponentialBackOff()
		batch      *Batch                      = new(Batch)
		nextSeqNum uint64
	)

	err := json.Unmarshal(msg.Body, batch)
	if err != nil {
		return err
	}

	nextSeqNum = batch.Start

	log.Printf("Processing batch [%d,%d]", batch.Start, batch.End)

	client = &http.Client{}

	expBackoff.MaxElapsedTime = 20 * time.Minute
	expBackoff.MaxInterval = time.Minute
	expBackoff.InitialInterval = time.Second

	backoffLogger := func(e error, wait time.Duration) {
		log.Printf("Failed requesting data: %v, waiting %s", e, wait)
	}

	requestExecutor := func() error {
		var err error
		resp, err = client.Do(req)

		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("Server returned status %s", resp.Status)
		}

		return nil
	}

	for {
		req = mf.api.Request(steam.GetMatchHistoryBySequenceNum, steam.StartAtMatchSeqNum(nextSeqNum))

		expBackoff.Reset()
		err = backoff.RetryNotify(requestExecutor, expBackoff, backoffLogger)
		if err != nil {
			return err
		}

		log.Printf("Response: %s", resp.Status)

		mhr, err := mf.getResponseBody(resp)
		if err != nil {
			return err
		}

		for _, match := range mhr.Matches {

			// only index matches that are in our batch
			if match.MatchSeqNum >= batch.End {
				return nil
			}

			nextSeqNum = match.MatchSeqNum + 1

			// submit the match
			data, err := msgpack.Marshal(match)
			if err != nil {
				return err
			}

			err = mf.producer.Publish("matches", data)
			if err != nil {
				return err
			}
		}

		// be nice with the API
		time.Sleep(time.Second)
	}

	panic("We shouldn't be getting here")
}

func (mf *MatchFetcher) getResponseBody(resp *http.Response) (*steam.MatchHistoryResult, error) {
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseData := new(steam.MatchHistoryResponse)

	err = json.Unmarshal(data, responseData)
	if err != nil {
		return nil, err
	}

	return responseData.Result, nil
}

var mf *MatchFetcher

func main() {
	var err error

	mf, err = NewMatchFetcher(MatchFetcherConfig{
		SteamApiKey: steam_api_key,
		NSQDAddress: "localhost:4150",
	})

	if err != nil {
		log.Fatal(err)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)

	<-sig
	log.Printf("Received signal, exitting")

	//bot.HttpClient = &debugHttpClient{http.DefaultClient}
}
