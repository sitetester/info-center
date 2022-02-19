package route

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/sitetester/info-center/controller"
	"github.com/sitetester/info-center/service"
	"net/http"
)

const serviceName = "info-center"

func SetupRouter() *gin.Engine {
	engine := gin.Default()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	engine.Use(gin.Recovery())

	apiService := service.NewApiService(redis.NewClient(&redis.Options{Addr: "redis:6379"}))
	apiController := controller.NewApiController(apiService)

	engine.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, fmt.Sprintf("[%s] API is functional", serviceName))
	})

	topicPath := fmt.Sprintf("/%s/:topic", serviceName)

	engine.POST(topicPath, apiController.HandleTopicPostRoute)
	engine.GET(topicPath, apiController.HandleTopicGetRoute)

	return engine
}
