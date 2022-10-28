package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/nikchis/grpc-vs-nng/internal/inng"
)

func main() {
	log.SetFlags(log.LUTC | log.Lmicroseconds | log.Lshortfile)

	fAddress := flag.String("a", "tcp://0.0.0.0:50050", "server address")
	flag.Parse()

	nsrv := inng.NewServer(*fAddress)

	if err := nsrv.StartListenAndServe(); err != nil {
		log.Fatalf("Failed to start listen and serve: %v", err)
	}

	sigInterrupt := make(chan os.Signal)
	signal.Notify(sigInterrupt, os.Interrupt)

	<-sigInterrupt
	nsrv.Close()
}
