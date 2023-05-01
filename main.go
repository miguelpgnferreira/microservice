package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/miguelpgnferreira/microservice/client"
	"github.com/miguelpgnferreira/microservice/proto"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		jsonAddr = flag.String("json", ":3000", "listen address the json transport")
		grpcAddr = flag.String("grpc", ":4000", "listen address the grpc transport")
		svc      = loggingService{priceService{}}
		ctx      = context.Background()
	)

	flag.Parse()

	grpcClient, err := client.NewGRPCClient(":4000")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			time.Sleep(3 * time.Second)
			resp, err := grpcClient.FetchPrice(ctx, &proto.PriceRequest{Ticker: "BTC"})
			if err != nil {
				log.Fatal(err)
			}

			fmt.Print("\n", resp)
		}
	}()

	go makeGRPCServerAndRun(*grpcAddr, svc)

	jsonServer := NewJSONAPIServer(*jsonAddr, svc)
	jsonServer.Run()

}
