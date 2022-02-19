package service

import (
	"encoding/json"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"log"
	"net/http"
	"time"
)

// maxAllowedConnectionTime Service should disconnect all the clients if they were consuming the stream for more than max allowed time (e.g. 30 sec).
const maxAllowedConnectionTime = 30

var id = 0

type RequestData struct {
	Topic string `valid:"alpha,required"`
	Msg   string `valid:"alpha,required"`
}

type Payload struct {
	Msg string
	Id  int
}

type ApiService struct {
	redisClient *redis.Client
}

func NewApiService(client *redis.Client) *ApiService {
	return &ApiService{redisClient: client}
}

func (s *ApiService) SaveMessage(topic string, msg string) error {
	rd := RequestData{Topic: topic, Msg: msg}
	if _, err := govalidator.ValidateStruct(rd); err != nil {
		return err
	}

	// Each message sent to the server should have unique auto-incrementing ID value.
	id += 1
	p := Payload{Msg: rd.Msg, Id: id}
	message, err := json.Marshal(p)
	if err != nil {
		return err
	}

	log.Printf("Publishing msg (`%s`) to topic (`%s`)", rd.Msg, rd.Topic)
	s.redisClient.Publish(topic, message)
	return nil
}

func (s *ApiService) GetMessages(rw gin.ResponseWriter, flusher http.Flusher, topic string) error {
	s.setHeaders(rw)

	connectedAt := time.Now()
	format := "event: timeout\ndata: %ds\n"

	chanMessage := s.redisClient.Subscribe(topic).Channel()

	for {
		select {
		case msg := <-chanMessage:
			p := &Payload{}
			if err := json.Unmarshal([]byte(msg.Payload), &p); err != nil {
				return err
			}

			if _, err := fmt.Fprintf(rw, "id: %d\nevent: %s\ndata: %s\n\n", p.Id, "msg", p.Msg); err != nil {
				return err
			}
			flusher.Flush()

			if s.isTimedOut(connectedAt) {
				if _, err := fmt.Fprintf(rw, format, maxAllowedConnectionTime); err != nil {
					return err
				}
				flusher.Flush()
				return nil
			}
		default:
			if s.isTimedOut(connectedAt) {
				if _, err := fmt.Fprintf(rw, format, maxAllowedConnectionTime); err != nil {
					return err
				}
				flusher.Flush()
				return nil
			}
		}
	}
}

// Set the headers related to event streaming.
func (s *ApiService) setHeaders(rw gin.ResponseWriter) {
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Transfer-Encoding", "chunked")
}

func (s *ApiService) isTimedOut(connectedAt time.Time) bool {
	diff := (int)(time.Now().Sub(connectedAt).Seconds())
	return diff > maxAllowedConnectionTime
}
