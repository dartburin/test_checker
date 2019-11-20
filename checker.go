package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

// Struct for result of request
type RetData struct {
	bad        bool
	timeOut    bool
	startTime  time.Time
	finishTime time.Time
}

// Struct for result of test
type Statistics struct {
	CntBad     uint
	CntTimeout uint
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

// Main function parsed initial parameters, created requests and check result
func main() {
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
	for i := uint(0); i < cntRepeat; i++ {
		go checkResource(domain, c, &client)
	}

	stat := Statistics{}
	countingStat(&stat, cntRepeat, c)

	// Print result of request test
	fmt.Println(stat)
}

// Check availability of resource
func checkResource(url *string, channel chan RetData, cl *http.Client) {
	ret := RetData{}

	ret.startTime = time.Now()
	resp, err := cl.Get(*url)
	ret.finishTime = time.Now()

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

	channel <- ret
}

//  Statistic counting
func countingStat(stat *Statistics, cnt uint, channel chan RetData) {

	len := time.Duration(0)
	l := RetData{}
	startTime := time.Now()
	finishTime := time.Time{}

	for i := uint(0); i < cnt; i++ {
		l = <-channel

		delta := l.finishTime.Sub(l.startTime)

		// Calc counts of not ok requests
		if l.bad {
			stat.CntBad++
		}

		if l.timeOut {
			stat.CntTimeout++
		}

		// Check times of a single request
		if delta < stat.MinTime || stat.MinTime == 0 {
			stat.MinTime = delta
		}

		if delta > stat.MaxTime {
			stat.MaxTime = delta
		}

		// Check times of all requests
		if l.startTime.Before(startTime) {
			startTime = l.startTime
		}

		if l.finishTime.After(finishTime) {
			finishTime = l.finishTime
		}

		len += delta
	}

	stat.FullTime = finishTime.Sub(startTime)
	stat.MiddleTime = len / time.Duration(cnt)
}
