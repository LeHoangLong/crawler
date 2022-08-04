package main

import (
	"context"
	"crawler_go/internal/controllers"
	"crawler_go/internal/repositories"
	"crawler_go/internal/server"
	"crawler_go/internal/services"
	"crawler_go/internal/views"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/playwright-community/playwright-go"
)

func main() {
	log.SetFlags(0)

	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

// run starts a http.Server for the passed in address
// with all requests handled by echoServer.
func run() error {
	l, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		return err
	}
	log.Printf("listening on http://%v", l.Addr())

	wsServer := server.MakeWsServer()

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	pageRepository := repositories.MakePageRepository(10, &browser)

	priceFetcherShopee := services.MakePriceFetcherServiceShopee(&pageRepository)
	priceFetcherLazada := services.MakePriceFetcherServiceLazada(&pageRepository)

	priceRequestController := controllers.MakePriceRequestController()
	priceRequestController.AddPriceFetcherService(&priceFetcherShopee)
	priceRequestController.AddPriceFetcherService(&priceFetcherLazada)
	priceRequestController.Start()
	defer priceRequestController.Stop()

	priceRquestView := views.MakePriceRequestViewWs(&priceRequestController)
	wsServer.RegisterEndpoint("price-request", &priceRquestView)
	s := &http.Server{
		Handler:      &wsServer,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	errc := make(chan error, 1)
	go func() {
		fmt.Println("serving server")
		errc <- s.Serve(l)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-errc:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return s.Shutdown(ctx)
}
