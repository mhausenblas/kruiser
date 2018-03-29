package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/mhausenblas/yages/yages"
	"google.golang.org/grpc"
)

func main() {
	serverAddr := flag.String("svc", "127.0.0.1:9000", "YAGES service address")
	flag.Parse()
	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := yages.NewEchoClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	res, err := c.Ping(ctx, &yages.Empty{})
	if err != nil {
		log.Fatal(err)
	}
	cancel()
	log.Println(res.Text)
}
