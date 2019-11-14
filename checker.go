package main

import (
    "fmt"
    "strings"
    "os"
    "strconv"
    "time"
    "net/http"
)

type RetData struct { // Result of Get 
    RetCode int
    Time time.Duration
}

type Statistics struct { // Result of Test
    CntBad int
    CntTimeout int
    MinTime time.Duration
    MaxTime time.Duration
    MiddleTime time.Duration
    FullTime time.Duration
}

func main() {
    //argsP := os.Args

    stat := Statistics{0,0,0,0,0,0}
    startTime := time.Now()

    args := os.Args[1:]  // Get argument list
    cntArgs := len(args) // Count of arguments
    timeout := 0

	if cntArgs < 2 { // Obligatory arguments not present
        if cntArgs == 1 { // Maybe help
            str := args[0]
            if strings.EqualFold(str, "-h") || strings.EqualFold(str, "-help") {
                fmt.Println("Help:")
                fmt.Println("    checker <domain name> <count of checks> [timeout (ms)]")
                fmt.Println("")
                os.Exit(0)        
            }
        }
        // It is error of arguments
		fmt.Println("Error: bad arguments")
		fmt.Println("    checker <domain name> <count of checks> [timeout (ms)]")
		fmt.Println("")
		os.Exit(100)
	} else if cntArgs >= 3 { // Maybe not obligatory argument present
        strTimeout := args[2]
        num, err := strconv.Atoi(strTimeout)

        if err == nil { // Ok not obligatory argument
            timeout = num
        } else { // Bad not obligatory argument
            fmt.Println("Warning: bad [timeout (ms)] argument. Ignored.")
            fmt.Println("")
        }
	}

    strCntRepeat := args[1] // Count repeat argument
    cntRepeat, err := strconv.Atoi(strCntRepeat)

    if err != nil { // Bad count repeat argument
		fmt.Println("Error: bad <count of checks> argument")
		fmt.Println("    checker <domain name> <count of checks> [timeout]")
		fmt.Println("")
		os.Exit(100)
    }


    domain := args[0] // Url argument

    // Print arguments
    fmt.Println("")
    fmt.Println("Check parameters:")
    fmt.Printf("Domain: %v\n", domain)
    fmt.Printf("Repeat count: %v\n", cntRepeat)

    if timeout > 0 {
        fmt.Printf("Timeout: %v\n", timeout)
    }
    fmt.Println("")




    // Start of test
    c := make(chan RetData)

    for i := 0; i < cntRepeat; i++ {
        go checkResource(domain, c)
    }




    // Result preparing
    checkTime := time.Second * time.Duration(timeout) / time.Duration(1000)
//    for l := range c {
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
    }


    finishTime := time.Now()
    stat.FullTime = finishTime.Sub(startTime)

    stat.MiddleTime = stat.FullTime / time.Duration(cntRepeat)


    // Print result
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


func checkResource(url string, channel chan RetData) {
    ret := RetData{0, 0}

//fmt.Println(" -- check 1")
    start := time.Now()
    client := http.Client{} 

    resp, err := client.Get(url) 
    if err != nil { 
        ret.RetCode = 404
    } else {
        defer resp.Body.Close() 
        ret.RetCode = resp.StatusCode
    }

    finish := time.Now()
    ret.Time = finish.Sub(start)
//fmt.Printf(" -- retCode %v\n", ret.RetCode)

    channel <- ret
//fmt.Println(" -- check 2")
    return
}