package main

import (
	"SimPro/app"
	"embed"
	"log"
)

//go:embed assets/*
var assetsFs embed.FS

func main() {
	application := app.NewApp(assetsFs)

	// 运行应用程序
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
