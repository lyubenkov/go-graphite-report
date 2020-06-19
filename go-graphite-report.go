package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lyubenkov/go-graphite-report/internal"
)

var (
	junitFilename = flag.String("f", "", "specify junit report filename")
	host          = flag.String("h", "", "specify graphite host address")
	port          = flag.Int("p", 0, "specify graphite port")
	prefix        = flag.String("x", "", "specify prefix for all metrics")
)

func main() {
	flag.Parse()

	if flag.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "%s does not accept positional arguments\n", os.Args[0])
		flag.Usage()
		os.Exit(1)
	}

	// Read junit report file
	report, err := os.Open(*junitFilename)
	if err != nil {
		fmt.Printf("Can't open junit report file: %s\n", err)
		os.Exit(1)
	}
	suites, err := internal.ReadJunitReport(report)
	if err != nil {
		fmt.Printf("Error reading from junit report file: %s\n", err)
		os.Exit(1)
	}

    // Map report to Graphite format
    metrics, err := internal.MapToGraphiteFormat(suites)
    if err != nil {
		fmt.Printf("Can't parse test result value: %s", err)
		os.Exit(1)
	}

	// Send metrics to Graphite
	err = internal.SendToGraphite(*host, *port, *prefix, metrics)
	if err != nil {
		fmt.Printf("Error sending metrics: %v", err)
		os.Exit(1)
	}
}
