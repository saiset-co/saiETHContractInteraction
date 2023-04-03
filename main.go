package main

import (
	"fmt"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/iamthe1whoknocks/saiEthInteraction/internal"
	"github.com/saiset-co/saiService"
)

func main() {
	svc := saiService.NewService("saiEthInteraction")

	svc.RegisterConfig("config.yml")

	is := internal.InternalService{Context: svc.Context}

	svc.RegisterInitTask(is.Init)

	svc.RegisterHandlers(
		is.NewHandler())

	if svc.GetConfig("common.enable_profiling", true).(bool) {
		mr := gin.Default()
		pprof.Register(mr)
		go mr.Run(fmt.Sprintf(":%d", svc.GetConfig("common.profiling_port", 8081).(int)))
	}

	svc.Start()

}
