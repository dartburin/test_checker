package main

import (
    "fmt"
    "flag"
    "time"
    "net/http"
)

/**
 @struct RetData Contained result of request
 */
type RetData struct {
    RetCode int
    Time time.Duration
}

/**
 @struct Statistics Contained result of test
 */
type Statistics struct {
    CntBad int
    CntTimeout int
    MinTime time.Duration
    MaxTime time.Duration
    MiddleTime time.Duration
    FullTime time.Duration
}

/**
 @brief Main function of checker
 @detailed This function parsed initial parameters, created requests, check result and
 calculated statistics
 */
func main() {
    stat := Statistics{}
    startTime := time.Now()

    var domain string
    flag.StringVar(&domain, "domain", "http://", "a domain name")

    var cntRepeat int
    flag.IntVar(&cntRepeat, "cnt", 4, "repeat requests count")

    var timeout int
    flag.IntVar(&timeout, "timeout", 1000000, "[timeout (ms)]")
    flag.Parse()

    /**
     Print arguments
     */
    fmt.Println("")
    fmt.Println("Check parameters:")
    fmt.Printf("Domain: %v\n", domain)
    fmt.Printf("Repeat count: %v\n", cntRepeat)

    if timeout > 1000000 {
        fmt.Printf("Timeout: %v\n", timeout)
    }
    fmt.Println("")



    /**
     Start of test
     */
    c := make(chan RetData)
    client := http.Client{} 

    for i := 0; i < cntRepeat; i++ {
        go checkResource(domain, c, &client)
    }




    /**
     Result preparing
     */
    checkTime := time.Second * time.Duration(timeout) / time.Duration(1000)
    len := time.Duration(0)
    l := RetData{}

    for i := 0; i < cntRepeat; i++ {
        l = <- c

        if l.RetCode < 200 || l.RetCode > 299 {
            stat.CntBad++
            continue
        }

        if timeout > 0 && l.Time > checkTime {
            stat.CntTimeout++
        }

        if l.Time < stat.MinTime  || stat.MinTime == 0 {
            stat.MinTime = l.Time
        }

        if l.Time > stat.MaxTime {
            stat.MaxTime = l.Time
        }

        len += l.Time
    }


    stat.FullTime = time.Since(startTime)
    if cntRepeat - stat.CntBad != 0 {
        stat.MiddleTime = len / time.Duration(cntRepeat - stat.CntBad)
    } else {
        stat.MiddleTime = 0
    }


    /**
     Print result
     */
    fmt.Println("")
    fmt.Println("Result statistic:")
    fmt.Printf("CntBad: %v\n", stat.CntBad)
    fmt.Printf("CntTimeout: %v\n", stat.CntTimeout)
    fmt.Printf("MinTime: %v\n", stat.MinTime)
    fmt.Printf("MaxTime: %v\n", stat.MaxTime)
    fmt.Printf("MiddleTime: %v\n", stat.MiddleTime)
    fmt.Printf("FullTime: %v\n", stat.FullTime)

    fmt.Println("")
}

/**
 @brief Check availability of resource
 @param url Domain name
 @param channel Channel for exchange with main
 @param cl HTTP client
*/
func checkResource(url string, channel chan RetData, cl *http.Client) {
    ret := RetData{}
//fmt.Println(" -- check 1")
    start := time.Now()

    resp, err := cl.Get(url) 
    if err != nil { 
        ret.RetCode = 500
    } else {
        ret.RetCode = resp.StatusCode
        defer resp.Body.Close() 
    }

    ret.Time = time.Since(start)
//fmt.Printf(" -- retCode %v\n", ret.RetCode)

    channel <- ret
//fmt.Println(" -- check 2")
    return
}