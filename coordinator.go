package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/bitly/go-nsq"
	"github.com/cenkalti/backoff"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

const (
	//hardcoded for now
	max_seq_num = 1214017946

	batch_size = 1e5

	batch_timeout = 1 * time.Hour
)

var (
	statusCodeError = errors.New("Invalid HTTP status code")
)

/*
func matches(c web.C, w http.ResponseWriter, req *http.Request) {
	matches := make([]steam.Match, 0)
	mgs := getMongoDB(c)
	defer req.Body.Close()

	if err := json.NewDecoder(req.Body).Decode(&matches); err != nil {
		log.Fatal(err)
	}

}

func finish(c web.C, w http.ResponseWriter, req *http.Request) {
	batch := new(Batch)
	defer req.Body.Close()

	if err := json.NewDecoder(req.Body).Decode(batch); err != nil {
		log.Fatal(err)
	}

}
*/

type coordinator struct {
	mgo *mgo.Session
}

func (c *coordinator) closeBatch(batch *Batch) error {
	return c.mgo.DB("dota").C("batches").Update(bson.M{
		"_id":       batch.Id,
		"completed": time.Time{},
	}, batch)
}

func (c *coordinator) nextBatch() (batch *Batch, err error) {
	batch = new(Batch)

	// find the latest batch
	err = c.mgo.DB("dota").C("batches").Find(nil).Sort("-created").One(batch)
	newBatch := &Batch{
		Created: time.Now().UTC(),
		Id:      bson.NewObjectId(),
	}

	switch err {
	case mgo.ErrNotFound:
		log.Print("Creating first batch")
		// this is the first batch
		newBatch.Start = startAt
		newBatch.End = startAt + batch_size

	case nil:
		log.Printf("Continuing at %d", batch.End)
		// add a new batch after the latest one
		newBatch.Start = batch.End
		newBatch.End = batch.End + batch_size

	default:
		return
	}

	batch = newBatch

	err = c.mgo.DB("dota").C("batches").Insert(batch)

	return
}

type Batch struct {
	Id bson.ObjectId `json:"_id" bson:"_id,omitempty"`

	Start     uint64    `json:"start" bson:"start"`
	End       uint64    `json:"end" bson:"end"`
	Created   time.Time `json:"created" bson:"created"`
	Completed time.Time `json:"completed" bson:"completed"`
	Attempts  uint64    `json:"attempts" bson:"attempts"`
}

type Worker struct {
	Id     bson.ObjectId `bson:"_id,omitempty"`
	Secret string        `bson:"secret"`
}

const (
	coordinatorMiddlewareKey = "coordinator"
	workerIdHeader           = "X-Worker-Id"
	workerSecretHeader       = "X-Worker-Secret"
)

type coordinatorMiddleware struct {
	mgo *mgo.Session
	c   *web.C
	h   http.Handler
}

func (c *coordinatorMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// expect authentication headers
	id := req.Header.Get(workerIdHeader)
	secret := req.Header.Get(workerSecretHeader)
	worker := new(Worker)

	m := c.mgo

	err := m.DB("dota").C("worker").FindId(bson.ObjectIdHex(id)).One(worker)
	log.Printf("worker: %s secret: %s err: %v", id, secret, err)
	if err == mgo.ErrNotFound {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if subtle.ConstantTimeCompare([]byte(secret), []byte(worker.Secret)) != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	c.c.Env[coordinatorMiddlewareKey] = c
	c.h.ServeHTTP(w, req)
}

func CoordinatorMiddleware(m *mgo.Session) web.MiddlewareType {
	return func(c *web.C, h http.Handler) http.Handler {
		return &coordinatorMiddleware{
			mgo: m,
			c:   c,
			h:   h,
		}
	}
}

func getCoordinatorMiddleware(c web.C) *coordinatorMiddleware {
	return c.Env[coordinatorMiddlewareKey].(*coordinatorMiddleware)
}

func getMongoDB(c web.C) *mgo.Session {
	return getCoordinatorMiddleware(c).mgo
}

type onDemandProducer struct {
	nsqdHttpAddress string
	handler         func(Stats)
	frequency       time.Duration
}

func NewOnDemandProducer(nsqdHttpAddress string, frequency time.Duration, handler func(Stats)) *onDemandProducer {
	if frequency == 0 {
		frequency = time.Second * 5
	}

	return &onDemandProducer{
		nsqdHttpAddress: nsqdHttpAddress,
		handler:         handler,
		frequency:       frequency,
	}
}

type StatsResponse struct {
	Data Stats
}

type Stats struct {
	Topics []TopicStats
}

func (s Stats) CombinedReadyCount(topicName, channelName string) (count int) {
	for _, topic := range s.Topics {
		if topic.Name != topicName {
			continue
		}

		for _, channel := range topic.Channels {
			if channel.Name != channelName {
				continue
			}

			for _, client := range channel.Clients {
				count += client.ReadyCount
			}

			count -= channel.Depth + channel.DeferredCount
		}
	}

	return
}

type TopicStats struct {
	Name     string `json:"topic_name"`
	Channels []ChannelStats
	//Depth    int
}

type ChannelStats struct {
	Name          string `json:"channel_name"`
	Depth         int
	DeferredCount int `json:"deferred_count"`
	Clients       []ClientStats
}

type ClientStats struct {
	ReadyCount int `json:"ready_count"`
}

func (n *onDemandProducer) checkDemand() error {
	resp, err := http.Get(n.nsqdHttpAddress + "/stats?format=json")

	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		return statusCodeError
	}
	defer resp.Body.Close()

	stats := new(StatsResponse)
	err = json.NewDecoder(resp.Body).Decode(stats)
	if err != nil {
		return err
	}

	n.handler(stats.Data)

	return nil
}

func (n *onDemandProducer) producerLoop() {
	for {
		if err := backoff.RetryNotify(n.checkDemand, backoff.NewExponentialBackOff(), func(e error, wait time.Duration) {
			log.Printf("Failed retrieving stats: %v, waiting %v", e, wait)
		}); err != nil {
			log.Fatal(err)
		}

		time.Sleep(n.frequency)
	}
}

func (n *onDemandProducer) Start() {
	go n.producerLoop()
}

var startAt uint64

func main() {
	mgs, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer mgs.Close()

	coord := &coordinator{mgs}

	stats := NewOnDemandProducer("http://localhost:4151", 0, func(s Stats) {
		rdy := s.CombinedReadyCount("batches", "worker")
		if rdy > 0 {
			cfg := nsq.NewConfig()
			prod, _ := nsq.NewProducer("localhost:4150", cfg)
			defer prod.Stop()

			for i := 0; i < rdy; i++ {

				batch, err := coord.nextBatch()
				if err != nil {
					log.Print(err)
				}
				data, _ := json.Marshal(batch)

				prod.Publish("batches", data)
			}
		}
	})

	stats.Start()

	if err := mgs.Ping(); err != nil {
		log.Fatal(err)
	}

	mgs.DB("dota").C("batches").EnsureIndex(mgo.Index{
		Key:    []string{"start"},
		Unique: true,
	})

	mgs.DB("dota").C("batches").EnsureIndexKey("created")
	mgs.DB("dota").C("batches").EnsureIndexKey("completed")

	startAt = 0

	goji.Use(CoordinatorMiddleware(mgs))
	//goji.Get("/auth", nsqAuth)
	goji.Serve()
}
