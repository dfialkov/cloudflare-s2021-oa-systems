package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

func main() {

	urlPointer := flag.String("url", "nourl", "The site's url")
	profilePointer := flag.Int("profile", 0, "Whether to run the profiler after making the request")
	helpPointer := flag.Bool("help", false, "Whether to display help")
	verbosePtr := flag.Bool("v", false, "Whether to run in verbose mode.")

	flag.Parse()

	if *helpPointer {
		fmt.Println("This is a simple HTTP GET client. To specify a URL, use the --url flag like --url=[your URL here]. The client does not work without a URL specified. To run a speed profiler, use the --profile flag. To display this menu again, use the --help flag. Please note that this client cannot send HTTPS requests.")
	}

	//How mandatory flags work in Go, I guess.
	if *urlPointer == "nourl" {
		fmt.Println("Please specify a URL")
		os.Exit(0)
	}
	if *profilePointer < 0 {
		fmt.Println("Negative profiler attempt values not allowed.")
		os.Exit(0)
	}

	if *verbosePtr {
		fmt.Println("Starting in verbose mode. Remove the -v flag to stop these messages.")
		fmt.Println("URL to test:", *urlPointer)
	}

	if *verbosePtr {
		fmt.Println("Dialing connection over tcp")
	}

	u, err := url.Parse(*urlPointer)
	if err != nil {
		println("Invalid URL. Exiting")
		os.Exit(0)
	}

	//Parsing in order of appearance in an html request
	queryHost := fmt.Sprintf("%s", u.Host)
	queryHost = fmt.Sprintf("%s%s", queryHost, ":80")

	conn, err := net.Dial("tcp", queryHost)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	if *verbosePtr {
		fmt.Println("Connection Established")
	}

	defer conn.Close()

	relPath := "/"
	if len(u.Path) > 0 {
		relPath = u.Path
	}

	finalQuery := fmt.Sprintf("GET %s HTTP/1.0\r\nHost: %s\r\n\r\n", relPath, u.Host)

	if *verbosePtr {
		fmt.Println(finalQuery)
	}

	fmt.Fprintf(conn, finalQuery)

	if *verbosePtr {
		fmt.Println("Request sent")
	}

	b, err := ioutil.ReadAll(conn)

	fmt.Printf("%s", b)

	if *profilePointer > 0 {
		if *verbosePtr {
			fmt.Println("Starting profiling")
		}

		conn.Close()
		var times []int
		var sizes []int
		var errors []string
		for i := 0; i < *profilePointer; i++ {
			conn.Close()
			start := time.Now()
			conn, _ := net.Dial("tcp", queryHost)
			fmt.Fprintf(conn, finalQuery)
			b, _ := ioutil.ReadAll(conn)
			duration := time.Since(start)
			responseString := string(b)
			respSize := len(b)
			sizes = append(sizes, respSize)
			times = append(times, int(duration.Milliseconds()))
			status := responseString[:strings.Index(responseString, "\n")]
			statusCode := status[9:12]
			if statusCode != "200" {
				errors = append(errors, status)
			}
		}
		conn.Close()
		//Now, analyze the data.
		fmt.Printf("Requests made: %d\n", *profilePointer)
		sum := 0
		//Sorting makes some of the values easier
		sort.Ints(times)
		//calc max and min simultaneously, why not
		for _, val := range times {

			sum += val
		}
		fmt.Printf("Fastest time: %d\n", times[0])
		fmt.Printf("Slowest time: %d\n", times[len(times)-1])
		fmt.Printf("Mean time: %d\n", sum/len(times))
		median := 0
		mNumber := len(times) / 2

		if len(times)%2 == 1 {
			median = times[mNumber]
		}

		median = (times[mNumber-1] + times[mNumber]) / 2
		fmt.Printf("Median time: %d\n", median)
		successPercent := 0
		if len(errors) == 0 {
			successPercent = 100
		} else {
			successPercent = *profilePointer * 100 / len(errors) * 100
		}

		fmt.Printf("Request Success Percentage: %d %%\n", successPercent)
		if len(errors) > 0 {
			fmt.Printf("Non-success error codes: \n")
		}

		for _, val := range errors {
			fmt.Printf("%s, ", val)
		}
		fmt.Println()
		maxSize := sizes[0]
		minSize := sizes[0]
		for _, val := range sizes {
			if val > maxSize {
				maxSize = val
			}
			if val < minSize {
				minSize = val
			}
		}
		fmt.Printf("Smallest response: %d bytes\n", minSize)
		fmt.Printf("Largest response: %d bytes\n", maxSize)
	}
}
