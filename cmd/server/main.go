package main

import (
	"log"

	"github.com/4strodev/hello_api/pkg/core/app"
	"github.com/4strodev/hello_api/pkg/features/auth"
	"github.com/4strodev/hello_api/pkg/shared"
	wiring "github.com/4strodev/wiring/pkg"
)

func main() {
	container := wiring.New()

	app := app.NewApp(container)
	app.AddAdapter(shared.NewFiberAdapter())
	app.AddComponent(&auth.AuthController{})
	log.Fatal(app.Start())
}
