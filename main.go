package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"runtime"
	"test-bench/config"
	"test-bench/handlers"
)

func main() {
	conf := config.GetConfig()

	goMaxProcs := os.Getenv("GOMAXPROCS")

	if goMaxProcs == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	r := gin.Default()

	r.GET("/sites", handlers.GetSites)

	if err := r.Run(fmt.Sprintf(":%d", conf.Port)); err != nil {
		fmt.Println(err)
	}
}
