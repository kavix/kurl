package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kavix/kurl/internal/igrpc"
)

func handleGRPCCommand(args []string) {
	if len(args) == 0 {
		fatal(fmt.Errorf("error: grpc requires a URL\nUsage: kurl grpc://<URL> [Method] [--list-services] [--proto <file>]"))
	}

	targetURL := ""
	method := ""
	listServices := false
	protoFile := ""
	var headers []string
	verbose := false
	insecure := false
	data := ""
	timeoutStr := ""
	cert := ""
	key := ""
	cacert := ""

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--list-services":
			listServices = true
		case arg == "--proto":
			val, next, err := takeValue(args, i)
			if err != nil {
				fatal(err)
			}
			protoFile = val
			i = next
		case strings.HasPrefix(arg, "--proto="):
			protoFile = strings.SplitN(arg, "=", 2)[1]
		case arg == "-H" || arg == "--header":
			val, next, err := takeValue(args, i)
			if err != nil {
				fatal(err)
			}
			headers = append(headers, val)
			i = next
		case strings.HasPrefix(arg, "-H=") || strings.HasPrefix(arg, "--header="):
			headers = append(headers, strings.SplitN(arg, "=", 2)[1])
		case arg == "-d" || arg == "--data":
			val, next, err := takeValue(args, i)
			if err != nil {
				fatal(err)
			}
			data = val
			i = next
		case strings.HasPrefix(arg, "-d=") || strings.HasPrefix(arg, "--data="):
			data = strings.SplitN(arg, "=", 2)[1]
		case arg == "-t" || arg == "--timeout":
			val, next, err := takeValue(args, i)
			if err != nil {
				fatal(err)
			}
			timeoutStr = val
			i = next
		case strings.HasPrefix(arg, "-t=") || strings.HasPrefix(arg, "--timeout="):
			timeoutStr = strings.SplitN(arg, "=", 2)[1]
		case arg == "-v" || arg == "--verbose":
			verbose = true
		case arg == "--cert":
			val, next, err := takeValue(args, i)
			if err != nil {
				fatal(err)
			}
			cert = val
			i = next
		case strings.HasPrefix(arg, "--cert="):
			cert = strings.SplitN(arg, "=", 2)[1]
		case arg == "--key":
			val, next, err := takeValue(args, i)
			if err != nil {
				fatal(err)
			}
			key = val
			i = next
		case strings.HasPrefix(arg, "--key="):
			key = strings.SplitN(arg, "=", 2)[1]
		case arg == "--cacert":
			val, next, err := takeValue(args, i)
			if err != nil {
				fatal(err)
			}
			cacert = val
			i = next
		case strings.HasPrefix(arg, "--cacert="):
			cacert = strings.SplitN(arg, "=", 2)[1]
		case arg == "-k" || arg == "--insecure":
			insecure = true
		case !strings.HasPrefix(arg, "-"):
			if targetURL == "" {
				targetURL = arg
			} else if method == "" {
				method = arg
			} else {
				fatal(fmt.Errorf("unexpected argument: %s", arg))
			}
		default:
			fatal(fmt.Errorf("unknown flag: %s", arg))
		}
	}

	if targetURL == "" {
		fatal(fmt.Errorf("error: missing gRPC endpoint URL"))
	}

	timeout := 30 * time.Second
	var err error
	if timeoutStr != "" {
		timeout, err = parseTimeout(timeoutStr)
		if err != nil {
			fatal(err)
		}
	}

	ctx := context.Background()
	err = igrpc.Run(ctx, igrpc.Options{
		URL:          targetURL,
		Method:       method,
		ListServices: listServices,
		ProtoFile:    protoFile,
		Data:         data,
		Headers:      headers,
		Timeout:      timeout,
		Verbose:      verbose,
		Insecure:     insecure,
		Cert:         cert,
		Key:          key,
		CACert:       cacert,
	})
	if err != nil {
		fatal(err)
	}
}
