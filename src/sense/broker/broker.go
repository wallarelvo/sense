package main

import (
    // config "../config"
    util "../util"
    "flag"
    "fmt"
    "net/http"
)

type LoadData struct {
    Timestamp int64
    Frequency float64
    DeltaFreq int
}

var heartbeatMap map[string]LoadData = make(map[string]LoadData)
var loadMap map[string]int = make(map[string]int)

//returns the address of the worker with minimum load
func getMinLoad() string {
    var minLoadData *LoadData = nil
    var minAddr *string = nil

    for addr_key, load_data := range heartbeatMap {
        if minLoadData == nil {
            minLoadData = new(LoadData)
            *minLoadData = load_data

            minAddr = new(string)
            *minAddr = addr_key
        }

        if load_data.DeltaFreq < minLoadData.DeltaFreq {
            *minLoadData = load_data
            *minAddr = addr_key
        }
    }

    // load_data := heartbeatMap[*minAddr]
    // load_data.Frequency++
    // heartbeatMap[*minAddr] = load_data

    return *minAddr
}

//returns the url witht the minimum load
func GetURL(w http.ResponseWriter, r *http.Request) {
    if len(heartbeatMap) < 1 {
        panic("No workers")
    }

    minAddr := getMinLoad()

    ld := heartbeatMap[minAddr]
    fmt.Fprint(w, util.Response{"address": minAddr, "count": ld.Frequency})
}

//routes the service producer to a worker
func MobileDeviceReroute(w http.ResponseWriter, r *http.Request) {
    if len(heartbeatMap) < 1 {
        panic("No workers")
    }

    minAddr := getMinLoad()

    http.Redirect(w, r, minAddr+"/producer.html", http.StatusFound)
}

//routes the service consumer to a worker
func ClientReroute(w http.ResponseWriter, r *http.Request) {
    if len(heartbeatMap) < 1 {
        panic("No workers")
    }

    minAddr := getMinLoad()
    http.Redirect(
        w, r,
        minAddr+r.RequestURI,
        http.StatusFound,
    )
}

func main() {

    http.HandleFunc("/heartbeat", util.RouteWrapper(UpdateWorkers))
    http.HandleFunc("/register", util.RouteWrapper(MobileDeviceReroute))
    http.HandleFunc("/url", util.RouteWrapper(GetURL))
    http.HandleFunc("/client/", util.RouteWrapper(ClientReroute))

    var addr_flag = flag.String(
        "addr",
        "localhost",
        "Address the http server binds to",
    )

    var port_flag = flag.String(
        "port",
        "8000",
        "Port used for http server",
    )

    flag.Parse()

    go removeDeadWorkersLoop(heartbeatMap)

    err := http.ListenAndServe(*addr_flag+":"+*port_flag, nil)
    fmt.Println(err)
}
