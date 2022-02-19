package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sitetester/info-center/service"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ApiController struct {
	service *service.ApiService
}

func NewApiController(apiService *service.ApiService) *ApiController {
	return &ApiController{service: apiService}
}

func (c *ApiController) HandleTopicPostRoute(ctx *gin.Context) {
	topic := ctx.Param("topic")
	msg := ctx.PostForm("msg")

	err := c.service.SaveMessage(topic, msg)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.Status(204)
}

func (c *ApiController) HandleTopicGetRoute(ctx *gin.Context) {
	topic := ctx.Param("topic")

	rw := ctx.Writer
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	err := c.service.GetMessages(rw, flusher, topic)
	if err != nil {
		log.Printf("Error in streaming topic messages: %s", err.Error())
		fmt.Fprintf(rw, err.Error())
	}
}
