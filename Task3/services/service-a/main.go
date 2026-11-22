package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdown, err := setupTracing(ctx, envOr("OTEL_SERVICE_NAME", "service-a"))
	if err != nil {
		log.Fatalf("otel: %v", err)
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err = shutdown(c); err != nil {
			log.Printf("otel shutdown: %v", err)
		}
	}()

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   10 * time.Second,
	}
	serviceBURL := envOr("SERVICE_B_URL", "http://service-b:8080")

	mux := http.NewServeMux()
	mux.Handle("/", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRoot(w, r, client, serviceBURL)
	}), "GET /"))

	srv := &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		log.Println("service-a listening on :8080")

		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(c)
}

func handleRoot(w http.ResponseWriter, r *http.Request, client *http.Client, serviceBURL string) {
	orderID := envOr("DEMO_ORDER_ID", "order-demo-001")
	span := trace.SpanFromContext(r.Context())
	span.SetAttributes(
		attribute.String("order.id", orderID),
		attribute.String("order.status", "CALC_REQUESTED"),
	)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, serviceBURL+"/", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("X-Order-Id", orderID)

	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("service-b call failed: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	span.SetAttributes(attribute.String("order.status", "PRICE_CALCULATED"))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(resp.StatusCode)
	_, _ = fmt.Fprintf(w, "service-a -> service-b\n%s", body)
}

func setupTracing(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	endpoint := envOr("OTEL_EXPORTER_OTLP_ENDPOINT", "simplest-collector:4317")

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// NewSchemaless: Default() уже со своей Schema URL — Merge с v1.24 падал с conflicting Schema URL
	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
