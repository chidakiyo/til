package main

import (
	"context"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"errors"
	"fmt"
	"github.com/favclip/ucon"
	"github.com/vvakame/til/appengine/go111-internal-grpc/echopb"
	"github.com/vvakame/til/appengine/go111-internal-grpc/log"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	rlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

var echoCli echopb.EchoClient
var echoOnce sync.Once

func main() {
	close, err := log.Init()
	if err != nil {
		rlog.Fatalf("Failed to create logger: %v", err)
	}
	defer close()

	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	})
	if err != nil {
		rlog.Fatalf("Failed to create stackdriver exporter: %v", err)
	}
	trace.RegisterExporter(exporter)
	view.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})
	defer exporter.Flush()

	ucon.Middleware(func(b *ucon.Bubble) error {
		b.Context = log.WithContext(b.Context, b.R)
		b.R = b.R.WithContext(b.Context)
		return b.Next()
	})
	ucon.Orthodox()

	ucon.HandleFunc("POST", "/echo", echoHandler)
	ucon.HandleFunc("GET", "/", indexHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		rlog.Printf("Defaulting to port %s for HTTP", port)
	}

	rlog.Printf("Listening HTTP on port %s", port)

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
		Handler: &ochttp.Handler{
			Handler:          ucon.DefaultMux,
			IsPublicEndpoint: true,
			Propagation:      &propagation.HTTPFormat{},
		},
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			rlog.Fatal(err)
		}
	}()

	if err := view.Register(ocgrpc.DefaultClientViews...); err != nil {
		rlog.Fatalf("Failed to register gRPC client views: %v", err)
	}
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		rlog.Fatalf("Failed to register gRPC server views: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	echopb.RegisterEchoServer(grpcServer, &echoServiceImpl{})
	reflection.Register(grpcServer)

	socketPath := filepath.Join(os.TempDir(), "go111-internal-grpc")
	os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		rlog.Fatal(err)
	}
	defer listener.Close()

	rlog.Printf("Listening gRPC on unix domain socket %s", socketPath)

	go func() {
		if err := grpcServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			rlog.Fatal(err)
		}
	}()

	rlog.Printf("running...")

	// setup graceful shutdown...
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	<-sigCh

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		rlog.Fatalf("graceful shutdown failure: %s", err)
	}
	grpcServer.GracefulStop()
	rlog.Printf("graceful shutdown successfully")
}

// indexHandler responds to requests with our greeting.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()

	ctx, span := trace.StartSpan(ctx, "indexHandler")
	defer span.End()

	log.Debugf(ctx, "Hi, 1")
	log.Infof(ctx, "Hi, 2")

	fmt.Fprint(w, "Hello, World!")
}

func echoHandler(ctx context.Context, req *echopb.SayRequest) (*echopb.SayResponse, error) {
	ctx, span := trace.StartSpan(ctx, "echoHandler")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("messageId", req.MessageId))
	span.AddAttributes(trace.StringAttribute("messageBody", req.MessageBody))

	echoOnce.Do(func() {
		socketPath := filepath.Join(os.TempDir(), "go111-internal-grpc")
		conn, err := grpc.Dial(
			socketPath,
			grpc.WithInsecure(),
			grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
				return net.DialTimeout("unix", addr, timeout)
			}),
			grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		)
		if err != nil {
			return
		}

		echoCli = echopb.NewEchoClient(conn)
	})
	if echoCli == nil {
		return nil, errors.New("echoCli is nil")
	}

	resp, err := echoCli.Say(ctx, req)
	status := status.Convert(err)
	log.Debugf(ctx, "%#v", status)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
