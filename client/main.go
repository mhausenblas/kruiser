package main

import (
	"context"
	"flag"
	"log"

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
	res, err := c.Ping(context.Background(), &yages.Empty{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res.Text)
}
