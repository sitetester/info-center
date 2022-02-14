package main

import (
	"github.com/sitetester/infocenter_redis/route"
)

func main() {
	engine := route.SetupRouter()
	err := engine.Run(":8081")
	if err != nil {
		panic(err)
	}
}
