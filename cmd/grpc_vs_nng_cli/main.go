package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/nikchis/grpc-vs-nng/internal/igrpc"
	"github.com/nikchis/grpc-vs-nng/internal/inng"
)

func main() {
	log.SetFlags(log.LUTC | log.Lmicroseconds | log.Lshortfile)

	fAddressGrpc := flag.String("g", "127.0.0.1:50051", "gRPC server address")
	fAddressNng := flag.String("n", "tcp://127.0.0.1:50050", "NNG server address")
	fFilepath := flag.String("f", "../etc/data/photo_01.jpg", "filepath")
	fCount := flag.Int("c", 10000, "count of testing calls")

	flag.Parse()

	payload, err := os.ReadFile(*fFilepath)
	if err != nil {
		log.Printf("Failed to read file on [%s]: %v", *fFilepath, err)
		log.Printf("Will use empty payload")
	}

	gcl, err := igrpc.NewClient(*fAddressGrpc)
	if err != nil {
		log.Fatalf("Failed to initiate gRPC client: %v", err)
	}
	defer gcl.Close()

	log.Printf("Start to make %d gRPC calls", *fCount)
	start := time.Now()
	for i := 0; i < *fCount; i++ {
		req := &igrpc.RequestImageProcessing{
			MessageId: strconv.Itoa(i),
			Left:      100,
			Bottom:    550,
			Right:     500,
			Top:       200,
			Quality:   92,
			Payload:   payload,
		}

		_, err := gcl.CallImageProcessing(req)
		if err != nil {
			log.Fatalf("Failed to make gRPC call: %v", err)
		}
	}
	log.Printf("Finished. Time: %6.4f sec", time.Since(start).Seconds())

	ncl, err := inng.NewClient(*fAddressNng)
	if err != nil {
		log.Fatalf("Failed to initiate NNG client: %v", err)
	}
	defer ncl.Close()

	log.Printf("Start to make %d NNG calls", *fCount)
	start = time.Now()
	for i := 0; i < *fCount; i++ {
		req := &inng.RequestImageProcessing{
			MessageId: strconv.Itoa(i),
			Left:      100,
			Bottom:    550,
			Right:     500,
			Top:       200,
			Quality:   92,
		}

		_, err := ncl.CallImageProcessing(req, payload)
		if err != nil {
			log.Fatalf("Failed to make NNG call: %v", err)
		}
	}
	log.Printf("Finished. Time: %6.4f sec", time.Since(start).Seconds())

}
