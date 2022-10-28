package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/nikchis/grpc-vs-nng/internal/igrpc"
)

func main() {
	log.SetFlags(log.LUTC | log.Lmicroseconds | log.Lshortfile)

	fPort := flag.String("p", "50051", "server port")
	flag.Parse()

	gsrv, err := igrpc.NewServer(*fPort)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	if err := gsrv.StartListenAndServe(); err != nil {
		log.Fatalf("Failed to start listen and serve: %v", err)
	}

	sigInterrupt := make(chan os.Signal)
	signal.Notify(sigInterrupt, os.Interrupt)

	<-sigInterrupt
	gsrv.Close()
}
