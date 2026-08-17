package igrpc

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Options struct {
	URL          string
	Method       string
	ListServices bool
	ProtoFile    string
	Data         string
	Headers      []string
	Timeout      time.Duration
	Verbose      bool
	Insecure     bool
	Cert         string
	Key          string
	CACert       string
}

func Run(ctx context.Context, opts Options) error {
	target := opts.URL
	var isTLS bool
	if strings.HasPrefix(target, "grpcs://") {
		isTLS = true
		target = strings.TrimPrefix(target, "grpcs://")
	} else if strings.HasPrefix(target, "grpc://") {
		isTLS = false
		target = strings.TrimPrefix(target, "grpc://")
	} else {
		// Default to insecure if no scheme but this package expects one usually
		isTLS = false
	}

	var creds credentials.TransportCredentials
	if isTLS {
		var err error
		creds, err = grpcurl.ClientTransportCredentials(opts.Insecure, opts.CACert, opts.Cert, opts.Key)
		if err != nil {
			return fmt.Errorf("failed to configure TLS credentials: %w", err)
		}
	} else {
		creds = insecure.NewCredentials()
	}

	dialCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	cc, err := grpc.DialContext(dialCtx, target, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer cc.Close()

	var descSource grpcurl.DescriptorSource

	if opts.ProtoFile != "" {
		descSource, err = grpcurl.DescriptorSourceFromProtoFiles(nil, opts.ProtoFile)
		if err != nil {
			return fmt.Errorf("failed to process proto file: %w", err)
		}
	} else {
		refClient := grpcreflect.NewClientAuto(dialCtx, cc)
		defer refClient.Reset()
		descSource = grpcurl.DescriptorSourceFromServer(dialCtx, refClient)
	}

	if opts.ListServices {
		services, err := grpcurl.ListServices(descSource)
		if err != nil {
			return fmt.Errorf("failed to list services: %w", err)
		}
		for _, s := range services {
			fmt.Println(s)
		}
		return nil
	}

	if opts.Method == "" {
		return fmt.Errorf("missing method to invoke")
	}

	var in io.Reader
	if opts.Data != "" {
		in = strings.NewReader(opts.Data)
	} else {
		// Empty JSON input if none provided for unary
		in = strings.NewReader("")
	}

	rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.FormatJSON, descSource, in, grpcurl.FormatOptions{
		EmitJSONDefaultFields: true,
		IncludeTextSeparator:  true,
		AllowUnknownFields:    true,
	})
	if err != nil {
		return fmt.Errorf("failed to construct parser and formatter: %w", err)
	}

	h := &grpcurl.DefaultEventHandler{
		Out:            os.Stdout,
		Formatter:      formatter,
		VerbosityLevel: 0,
	}
	if opts.Verbose {
		h.VerbosityLevel = 1
	}

	err = grpcurl.InvokeRPC(ctx, descSource, cc, opts.Method, opts.Headers, h, rf.Next)
	if err != nil {
		return fmt.Errorf("rpc error: %w", err)
	}

	if h.Status.Code() != 0 {
		grpcurl.PrintStatus(os.Stderr, h.Status, formatter)
	}

	return nil
}
