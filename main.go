package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"brew-terminal-curl/client"
	"brew-terminal-curl/color"
	"brew-terminal-curl/printer"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type cliOptions struct {
	method       string
	url          string
	data         string
	headers      []string
	timeout      time.Duration
	noColor      bool
	headersOnly  bool
	bodyOnly     bool
	raw          bool
	verbose      bool
	outputPath   string
	showHelp     bool
	showVersion  bool
	installAlias bool
}

func main() {
	opts, err := parseCLI(os.Args[1:])
	if err != nil {
		fatal(err)
	}
	if opts.showVersion {
		fmt.Printf("kurl version %s (commit: %s, built at: %s)\n", version, commit, date)
		return
	}
	if opts.showHelp {
		printUsage()
		return
	}
	if opts.installAlias {
		installShellAlias()
		return
	}

	useColor := color.AutoEnabled(os.Stdout) && !opts.noColor
	start := time.Now()

	result, err := client.Fetch(client.Options{
		Method:  opts.method,
		URL:     opts.url,
		Data:    []byte(opts.data),
		Headers: opts.headers,
		Timeout: opts.timeout,
		Verbose: opts.verbose,
	})
	if err != nil {
		fatal(err)
	}

	printerOptions := printer.Options{
		Color:       useColor,
		Raw:         opts.raw,
		HeadersOnly: opts.headersOnly,
		BodyOnly:    opts.bodyOnly,
		Verbose:     opts.verbose,
		OutputPath:  opts.outputPath,
	}

	bw := bufio.NewWriter(os.Stdout)
	if err := printer.Render(bw, result, printerOptions, time.Since(start)); err != nil {
		bw.Flush()
		fatal(err)
	}
	bw.Flush()
}

func parseCLI(args []string) (cliOptions, error) {
	options := cliOptions{method: "GET", timeout: 30 * time.Second}
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			options.showHelp = true
		case arg == "--install-alias":
			options.installAlias = true
		case arg == "--version" || arg == "-V":
			options.showVersion = true
		case arg == "-X" || arg == "--method":
			value, next, err := takeValue(args, i)
			if err != nil {
				return options, err
			}
			options.method = strings.ToUpper(value)
			i = next
		case strings.HasPrefix(arg, "-X=") || strings.HasPrefix(arg, "--method="):
			options.method = strings.ToUpper(strings.SplitN(arg, "=", 2)[1])
		case arg == "-d" || arg == "--data":
			value, next, err := takeValue(args, i)
			if err != nil {
				return options, err
			}
			options.data = value
			i = next
		case strings.HasPrefix(arg, "-d=") || strings.HasPrefix(arg, "--data="):
			options.data = strings.SplitN(arg, "=", 2)[1]
		case arg == "-H" || arg == "--header":
			value, next, err := takeValue(args, i)
			if err != nil {
				return options, err
			}
			options.headers = append(options.headers, value)
			i = next
		case strings.HasPrefix(arg, "-H=") || strings.HasPrefix(arg, "--header="):
			options.headers = append(options.headers, strings.SplitN(arg, "=", 2)[1])
		case arg == "-t" || arg == "--timeout":
			value, next, err := takeValue(args, i)
			if err != nil {
				return options, err
			}
			duration, err := time.ParseDuration(value + "s")
			if err != nil {
				return options, fmt.Errorf("invalid timeout %q: %w", value, err)
			}
			options.timeout = duration
			i = next
		case strings.HasPrefix(arg, "-t=") || strings.HasPrefix(arg, "--timeout="):
			value := strings.SplitN(arg, "=", 2)[1]
			duration, err := time.ParseDuration(value + "s")
			if err != nil {
				return options, fmt.Errorf("invalid timeout %q: %w", value, err)
			}
			options.timeout = duration
		case arg == "--no-color":
			options.noColor = true
		case arg == "--headers-only":
			options.headersOnly = true
		case arg == "--body-only":
			options.bodyOnly = true
		case arg == "--raw":
			options.raw = true
		case arg == "-v" || arg == "--verbose":
			options.verbose = true
		case arg == "-o" || arg == "--output":
			value, next, err := takeValue(args, i)
			if err != nil {
				return options, err
			}
			options.outputPath = value
			i = next
		case strings.HasPrefix(arg, "-o=") || strings.HasPrefix(arg, "--output="):
			options.outputPath = strings.SplitN(arg, "=", 2)[1]
		case strings.HasPrefix(arg, "-"):
			return options, fmt.Errorf("unknown flag %q", arg)
		default:
			positional = append(positional, arg)
		}
	}

	if options.showHelp {
		return options, nil
	}

	method, urlValue, err := resolveTarget(positional)
	if err != nil {
		return options, err
	}
	if len(positional) == 2 {
		options.method = method
	} else if options.method == "" || options.method == "GET" {
		options.method = method
	}
	options.url = urlValue

	if options.headersOnly && options.bodyOnly {
		return options, fmt.Errorf("--headers-only and --body-only cannot be used together")
	}

	return options, nil
}

func resolveTarget(positional []string) (string, string, error) {
	if len(positional) == 0 {
		return "", "", nil
	}
	if len(positional) == 1 {
		return "GET", positional[0], nil
	}
	if len(positional) > 2 {
		return "", "", fmt.Errorf("expected at most METHOD and URL")
	}

	if isMethodToken(positional[0]) {
		return strings.ToUpper(positional[0]), positional[1], nil
	}

	return "", "", fmt.Errorf("expected a METHOD and URL, or just a URL")
}

func looksLikeURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

func isMethodToken(value string) bool {
	if value == "" {
		return false
	}
	if strings.ContainsAny(value, "/.:?&=") {
		return false
	}
	return value == strings.ToUpper(value)
}

func takeValue(args []string, index int) (string, int, error) {
	if strings.Contains(args[index], "=") {
		return strings.SplitN(args[index], "=", 2)[1], index, nil
	}
	if index+1 >= len(args) {
		return "", index, fmt.Errorf("missing value for %q", args[index])
	}
	return args[index+1], index + 1, nil
}

func fatal(err error) {
	if err == nil {
		return
	}
	colored := color.ErrorText(color.AutoEnabled(os.Stderr), err.Error())
	fmt.Fprintln(os.Stderr, colored)
	os.Exit(1)
}

func printUsage() {
	fmt.Fprintln(os.Stdout, "kurl [METHOD] <URL> [flags]")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Flags:")
	fmt.Fprintln(os.Stdout, "  -X, --method      HTTP method (default GET)")
	fmt.Fprintln(os.Stdout, "  -d, --data        Request body")
	fmt.Fprintln(os.Stdout, "  -H, --header      Add header (repeatable)")
	fmt.Fprintln(os.Stdout, "  -t, --timeout     Timeout in seconds (default 30)")
	fmt.Fprintln(os.Stdout, "  --no-color        Disable color output")
	fmt.Fprintln(os.Stdout, "  --headers-only    Show only response headers")
	fmt.Fprintln(os.Stdout, "  --body-only       Show only response body")
	fmt.Fprintln(os.Stdout, "  --raw             Raw output, no formatting")
	fmt.Fprintln(os.Stdout, "  -v, --verbose     Show request info too")
	fmt.Fprintln(os.Stdout, "  -o, --output      Save body to file")
	fmt.Fprintln(os.Stdout, "  --install-alias   Install zsh/bash alias to prevent url globbing")
}

func installShellAlias() {
	home, err := os.UserHomeDir()
	if err != nil {
		fatal(err)
	}

	zshrcPath := home + "/.zshrc"
	aliasLine := `alias kurl="noglob kurl"`

	content, err := os.ReadFile(zshrcPath)
	if err != nil && !os.IsNotExist(err) {
		fatal(err)
	}

	if strings.Contains(string(content), aliasLine) {
		fmt.Println("Alias already exists in " + zshrcPath)
		return
	}

	f, err := os.OpenFile(zshrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fatal(err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n# Prevent zsh from breaking on URLs with ? and &\n" + aliasLine + "\n"); err != nil {
		fatal(err)
	}

	fmt.Println("✅ Successfully added 'noglob' alias to " + zshrcPath)
	fmt.Println("Please run 'source ~/.zshrc' or restart your terminal to apply the changes.")
}
