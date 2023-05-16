package main

import (
	"github.com/saiset-co/saiETHContractInteraction/internal"
	"github.com/saiset-co/saiService"
)

func main() {
	svc := saiService.NewService("saiEthInteraction")

	svc.RegisterConfig("config.yml")

	is := internal.InternalService{Context: svc.Context}

	svc.RegisterInitTask(is.Init)

	//defer is.Db.Close()

	svc.RegisterHandlers(
		is.NewHandler())

	svc.Start()

}
