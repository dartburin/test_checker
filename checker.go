package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

type RetData struct { // Result of request
	bad     bool
	timeOut bool
	Time    time.Duration
}

type Statistics struct { // Result of test
	CntBad     int
	CntTimeout int
	MinTime    time.Duration
	MaxTime    time.Duration
	MiddleTime time.Duration
	FullTime   time.Duration
}

// String is the Stringer interface of Statistics.
func (s Statistics) String() string {
	return fmt.Sprintf("\nResult statistic:\nCnt: bad=%v, timeout=%v\nTime: min=%v, middle=%v, max=%v, full=%v",
		s.CntBad, s.CntTimeout, s.MinTime, s.MiddleTime, s.MaxTime, s.FullTime)
}

// Main function parsed initial parameters, created requests, check result and
// calculated statistics
func main() {
	stat := Statistics{}
	startTime := time.Now()

	domain := flag.String("domain", "", "a domain name")

	var cntRepeat uint
	flag.UintVar(&cntRepeat, "cnt", 0, "repeat requests count - positive")

	var timeout uint
	flag.UintVar(&timeout, "timeout", 0, "[timeout (ms) - optional, positive]")
	flag.Parse()

	// Check existing obligatory params
	if *domain == "" || cntRepeat == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	c := make(chan RetData)
	client := http.Client{
		Timeout: time.Second * time.Duration(timeout) / time.Duration(1000),
	}

	//  Start gorutines with http requests
	var i uint
	for i = 0; i < cntRepeat; i++ {
		go checkResource(domain, c, &client)
	}

	//  Statistic counting
	len := time.Duration(0)
	l := RetData{}

	for i = 0; i < cntRepeat; i++ {
		l = <-c

		if l.bad {
			stat.CntBad++
		}

		if l.timeOut {
			stat.CntTimeout++
		}

		if l.Time < stat.MinTime || stat.MinTime == 0 {
			stat.MinTime = l.Time
		}

		if l.Time > stat.MaxTime {
			stat.MaxTime = l.Time
		}

		len += l.Time
	}

	stat.FullTime = time.Since(startTime)
	stat.MiddleTime = len / time.Duration(cntRepeat)

	fmt.Println(stat) // Print result of request test
}

// Check availability of resource
func checkResource(url *string, channel chan RetData, cl *http.Client) {
	ret := RetData{}
	start := time.Now()

	resp, err := cl.Get(*url)
	if err != nil {
		ret.bad = true
		if terr, ok := err.(net.Error); ok && terr.Timeout() {
			ret.timeOut = true
		}
	} else {
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			ret.bad = true
		}
		defer resp.Body.Close()
	}

	ret.Time = time.Since(start)
	channel <- ret
}
