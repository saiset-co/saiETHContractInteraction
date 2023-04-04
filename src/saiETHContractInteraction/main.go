package main

import (
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

	svc.Start()

}
