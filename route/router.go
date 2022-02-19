package route

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/sitetester/info-center/controller"
	"github.com/sitetester/info-center/service"
)

const serviceName = "info-center"

func SetupRouter() *gin.Engine {
	engine := gin.Default()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	engine.Use(gin.Recovery())

	apiService := service.NewApiService(redis.NewClient(&redis.Options{Addr: "redis:6379"}))
	apiController := controller.NewApiController(apiService)

	engine.GET("/", func(ctx *gin.Context) { ctx.String(200, "["+serviceName+"] API is functional!") })
	engine.POST("/"+serviceName+"/:topic", apiController.HandleTopicPostRoute)
	engine.GET("/"+serviceName+"/:topic", apiController.HandleTopicGetRoute)

	return engine
}
