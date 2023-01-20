package main

import (
	"log"
	"prom/app/config"
)

var conf = config.GetConfig()

func main() {
  a, err := initializeApplication()

  if err != nil {
    log.Fatal(err)
  }

	a.Start()
	a.Shutdown()
}
