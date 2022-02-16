package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"capactio/lambda/internal/generated"
	"capactio/lambda/internal/resolvers"

	"github.com/schartey/dgraph-lambda-go/api"
)

func main() {

	// Catch interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// WaitGroup for server shutdown
	wg := &sync.WaitGroup{}
	wg.Add(1)

	resolver := &resolvers.Resolver{}
	executer := generated.NewExecuter(resolver)
	lambda := api.New(executer)
	srv, err := lambda.Serve(wg)
	if err != nil {
		fmt.Println(err)
	}

	// Interrupt signal received
	<-c
	fmt.Println("Shutdown request (Ctrl-C) caught.")
	fmt.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println(err)
	}
	// Wait for server shutdown
	wg.Wait()

}
