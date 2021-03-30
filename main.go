package main

import (
    "context"
    "flag"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "os/signal"
    "strings"
    "sync"
    "time"

    "github.com/go-redis/redis/v8"
)

const (
    totalTime               = 10 * time.Second
    maxRoutines             = uint64(50)
    rltbKeyNameDefault      = "rltbDefault"
    maxReqsPerPeriod        = 200
    rateLimitPeriodCount    = 500
    rateLimitPeriodDuration = time.Millisecond
)

var (
    ctx = context.Background()
)

type counters struct {
    counter uint64
}

func getValueFromRedis(cc chan<- int, wg *sync.WaitGroup, rdb *redis.Client) {
    timeoutCtx, cf := context.WithTimeout(context.Background(), totalTime)
    defer cf()
    for {
        select {
        case <-timeoutCtx.Done():
            wg.Done()
            return
        default:
            v, err := rdb.Incr(context.Background(), rltbKeyNameDefault).Result()
            if err != nil {
                fmt.Printf("Error obtaining rate limiting bucket! %s\n", err.Error())
                _ = os.Stdout.Sync()
                continue
            }
            if uint64(v) > maxReqsPerPeriod {
                continue
            }
            resp, err := http.Get("http://localhost:3000/healthz")
            if err != nil {
                fmt.Printf("Error in req! %s\n", err.Error())
                _ = os.Stdout.Sync()
                continue
            }
            bs, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                fmt.Printf("Error reading body! %s\n", err.Error())
                _ = os.Stdout.Sync()
                continue
            }
            if !strings.Contains(string(bs), "ok") {
                fmt.Printf("Error not ok: %s!\n", string(bs))
                _ = os.Stdout.Sync()
                continue
            }
            cc <- 1
        }
    }
}

func resetFunc(bc <-chan bool, tc <-chan time.Time, rc *redis.Client, wg *sync.WaitGroup) {
    for {
        select {
        case <-bc:
            wg.Done()
            return
        case <-tc:
            fmt.Printf("Resetting at %s\n", time.Now().String())
            _ = os.Stdout.Sync()
            rc.Set(ctx, rltbKeyNameDefault, 0, 1*time.Second)
        }
    }
}

func handleCounting(cc <-chan int, c *counters) {
    for n := range cc {
        if n < 0 {
            break
        }
        c.counter++
    }
}

func main() {
    var (
        rdbOptions = &redis.Options{
            Addr:     "localhost:6379",
            Password: "",
            DB:       0,
        }
        rdb         = redis.NewClient(rdbOptions)
        counter     = &counters{counter: uint64(0)}
        counterChan = make(chan int, 1000)
        bucketChan  = make(chan bool)
        wg          = &sync.WaitGroup{}
        t1          time.Time
        resetter    bool
    )
    defer func() {
        _ = rdb.Close()
    }()
    defer close(counterChan)
    defer close(bucketChan)

    flag.BoolVar(&resetter, "resetter", false, "Start resetter")
    flag.Parse()

    if resetter {
        fmt.Printf("Press Ctrl+C to stop resetter.\n")
        reqIntChan := make(chan os.Signal, 1)
        signal.Notify(reqIntChan, os.Interrupt)
        ticker := time.NewTicker(rateLimitPeriodCount * rateLimitPeriodDuration)
        defer ticker.Stop()
        wg.Add(1)
        go resetFunc(bucketChan, ticker.C, rdb, wg)
        <-reqIntChan                 // we got a Ctrl+C.
        fmt.Printf("\nFinishing.\n") // let user know
        bucketChan <- true           // signal resetter that we're done.
        wg.Wait()                    // ensure routine finishes.
        fmt.Println("Done.")
        return
    }

    const waitSeconds = 1
    fmt.Printf("Waiting %d seconds for Redis client to properly connect ...\n", waitSeconds)
    time.Sleep(waitSeconds * time.Second)
    fmt.Printf("Starting ...\n")
    t1 = time.Now()

    for i := uint64(0); i < maxRoutines; i++ {
        go getValueFromRedis(counterChan, wg, rdb)
        wg.Add(1)
    }
    go handleCounting(counterChan, counter)

    wg.Wait()

    t2 := time.Now()

    counterChan <- -1
    time.Sleep(1 * time.Second)

    fmt.Println()
    fmt.Printf("Should have around %d requests in total.\n", maxReqsPerPeriod*(totalTime/(rateLimitPeriodCount*rateLimitPeriodDuration)))
    fmt.Printf("%d requests sent successfully in %0.4f seconds\n", counter.counter, float64(t2.Sub(t1).Milliseconds())/float64(1000))
    _ = os.Stdout.Sync()
}
