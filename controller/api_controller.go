package controller

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

type ErrorResponse struct {
	Error string
}

type ApiController struct {
	redisClient *redis.Client
}

func NewApiController(client *redis.Client) *ApiController {
	return &ApiController{redisClient: client}
}

func (c *ApiController) HandleTopicPostRoute(ctx *gin.Context) {
	topic := ctx.Param("topic")
	msg := ctx.PostForm("msg")

	rd := RequestData{
		Topic: topic,
		Msg:   msg,
	}

	if _, err := govalidator.ValidateStruct(rd); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Each message sent to the server should have unique auto-incrementing ID value.
	id += 1
	p := Payload{Msg: rd.Msg, Id: id}
	message, err := json.Marshal(p)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	log.Printf("Publishing msg (`%s`) to topic (`%s`)", rd.Msg, rd.Topic)
	c.redisClient.Publish(topic, message)
	ctx.Status(204)
}

func (c *ApiController) HandleTopicGetRoute(ctx *gin.Context) {
	connectedAt := time.Now()
	topic := ctx.Param("topic")

	rw := ctx.Writer
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Set the headers related to event streaming.
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Transfer-Encoding", "chunked")

	format := "event: timeout\ndata: %ds\n"
	chanMessage := c.redisClient.Subscribe(topic).Channel()

	for {
		select {
		case msg := <-chanMessage:

			p := &Payload{}
			err := json.Unmarshal([]byte(msg.Payload), &p)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Fprintf(rw, "id: %d\nevent: %s\ndata: %s\n\n", p.Id, "msg", p.Msg)
			flusher.Flush()

			if isTimedOut(connectedAt) {
				fmt.Fprintf(rw, format, maxAllowedConnectionTime)
				flusher.Flush()
				return
			}
		default:
			if isTimedOut(connectedAt) {
				fmt.Fprintf(rw, format, maxAllowedConnectionTime)
				flusher.Flush()
				return
			}
		}
	}
}

func isTimedOut(connectedAt time.Time) bool {
	diff := (int)(time.Now().Sub(connectedAt).Seconds())
	return diff > maxAllowedConnectionTime
}
