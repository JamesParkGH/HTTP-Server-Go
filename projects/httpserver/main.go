package main

import (
    "context"
    "errors"
    "fmt"
    "io"
    "net"
    "net/http"
)

const keyServerAddr = "serverAddr"

func getRoot(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    fmt.Printf("%s: got / request\n", ctx.Value(keyServerAddr))
    io.WriteString(w, "This is my website!\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    fmt.Printf("%s: got /hello request\n", ctx.Value(keyServerAddr))
    io.WriteString(w, "Hello, HTTP!\n")
}

func main() {
    // Create multiplexer
    mux := http.NewServeMux()
    mux.HandleFunc("/", getRoot)
    mux.HandleFunc("/hello", getHello)

    // Create a cancellable context
    ctx, cancelCtx := context.WithCancel(context.Background())

    // Configure first server - port 3333
    serverOne := &http.Server{
        Addr:    ":3333",
        Handler: mux,
        BaseContext: func(l net.Listener) context.Context {
            ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
            return ctx
        },
    }

    // Configure second server - port 4444
    serverTwo := &http.Server{
        Addr:    ":4444",
        Handler: mux,
        BaseContext: func(l net.Listener) context.Context {
            ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
            return ctx
        },
    }

    // Start first server in a goroutine
    go func() {
        fmt.Printf("Starting server at port 3333\n")
        err := serverOne.ListenAndServe()
        if errors.Is(err, http.ErrServerClosed) {
            fmt.Printf("server one closed\n")
        } else if err != nil {
            fmt.Printf("error listening for server one: %s\n", err)
        }
        cancelCtx()
    }()

    // Start second server in a goroutine
    go func() {
        fmt.Printf("Starting server at port 4444\n")
        err := serverTwo.ListenAndServe()
        if errors.Is(err, http.ErrServerClosed) {
            fmt.Printf("server two closed\n")
        } else if err != nil {
            fmt.Printf("error listening for server two: %s\n", err)
        }
        cancelCtx()
    }()

    // Keep program running until context is canceled
    <-ctx.Done()
}