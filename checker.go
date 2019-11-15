package main

import (
    "fmt"
    "os"
    "strconv"
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

    args := os.Args[1:]  //!< Get argument list
    cntArgs := len(args) //!< Count of arguments
    timeout := 0

	if cntArgs < 2 { 
        /** 
         Obligatory arguments not present
         It is error of arguments
         */
		fmt.Println("Error: bad arguments")
		fmt.Println("    checker <domain name> <count of checks> [timeout (ms)]")
		fmt.Println("")
		os.Exit(100)
	} else if cntArgs >= 3 { 
        /**
         Check not obligatory argument
         */
        strTimeout := args[2]
        num, err := strconv.Atoi(strTimeout)

        if err == nil {
            timeout = num
        } else {
            fmt.Println("Warning: bad [timeout (ms)] argument. Ignored.")
            fmt.Println("")
        }
	}

    strCntRepeat := args[1] //!< Count repeat argument
    cntRepeat, err := strconv.Atoi(strCntRepeat)
    /**
     Check count repeat argument
     */
    if err != nil { 
		fmt.Println("Error: bad <count of checks> argument")
		fmt.Println("    checker <domain name> <count of checks> [timeout]")
		fmt.Println("")
		os.Exit(100)
    }


    domain := args[0] //!< Url argument

    /**
     Print arguments
     */
    fmt.Println("")
    fmt.Println("Check parameters:")
    fmt.Printf("Domain: %v\n", domain)
    fmt.Printf("Repeat count: %v\n", cntRepeat)

    if timeout > 0 {
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
    stat.MiddleTime = len / time.Duration(cntRepeat - stat.CntBad)


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