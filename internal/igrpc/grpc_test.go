package igrpc_test

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kavix/kurl/internal/igrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func TestGRPCIntegration(t *testing.T) {
	// Start a local gRPC server
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// Register Health and Reflection services
	healthServer := health.NewServer()
	healthServer.SetServingStatus("MyService", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	reflection.Register(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
	defer s.Stop()

	targetURL := "grpc://" + lis.Addr().String()

	t.Run("ListServices", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		opts := igrpc.Options{
			URL:          targetURL,
			ListServices: true,
			Timeout:      5 * time.Second,
		}

		err := igrpc.Run(context.Background(), opts)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}

		w.Close()
		os.Stdout = oldStdout
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if !strings.Contains(output, "grpc.health.v1.Health") {
			t.Errorf("expected output to contain 'grpc.health.v1.Health', got: %s", output)
		}
	})

	t.Run("InvokeMethod", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		opts := igrpc.Options{
			URL:     targetURL,
			Method:  "grpc.health.v1.Health/Check",
			Data:    `{"service": "MyService"}`,
			Timeout: 5 * time.Second,
		}

		err := igrpc.Run(context.Background(), opts)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}

		w.Close()
		os.Stdout = oldStdout
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if !strings.Contains(output, "SERVING") {
			t.Errorf("expected output to contain 'SERVING', got: %s", output)
		}
	})
}

func BenchmarkGRPCInvocation(b *testing.B) {
	// Start a local gRPC server
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	healthServer := health.NewServer()
	healthServer.SetServingStatus("MyService", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	reflection.Register(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
	defer s.Stop()

	targetURL := "grpc://" + lis.Addr().String()

	b.Run("ReflectionInvocation", func(b *testing.B) {
		// Suppress output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		opts := igrpc.Options{
			URL:     targetURL,
			Method:  "grpc.health.v1.Health/Check",
			Data:    `{"service": "MyService"}`,
			Timeout: 5 * time.Second,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := igrpc.Run(context.Background(), opts)
			if err != nil {
				b.Fatalf("expected no error, got: %v", err)
			}
		}
		b.StopTimer()
		w.Close()
		os.Stdout = oldStdout
		io.Copy(io.Discard, r)
	})
}
