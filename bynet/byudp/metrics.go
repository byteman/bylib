package byudp

import (
	"bylib/bylog"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
)

var (
	handleExported *expvar.Int
	connExported   *expvar.Int
	timeExported   *expvar.Float
	qpsExported    *expvar.Float
)

//func init() {
//	handleExported = expvar.NewInt("TotalHandle")
//	connExported = expvar.NewInt("TotalConn")
//	timeExported = expvar.NewFloat("TotalTime")
//	qpsExported = expvar.NewFloat("QPS")
//}

// MonitorOn starts up an HTTP monitor on port.
func MonitorOn(port int) {
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			bylog.Error("%s",err)
			return
		}
	}()
}

func addTotalConn(delta int64) {
	connExported.Add(delta)
	calculateQPS()
}

func addTotalHandle() {
	handleExported.Add(1)
	calculateQPS()
}

func addTotalTime(seconds float64) {
	timeExported.Add(seconds)
	calculateQPS()
}

func calculateQPS() {
	totalConn, err := strconv.ParseInt(connExported.String(), 10, 64)
	if err != nil {
		bylog.Error("calculateQPS totalConn err=%s",err)
		return
	}

	totalTime, err := strconv.ParseFloat(timeExported.String(), 64)
	if err != nil {
		bylog.Error("calculateQPS totalTime err=%s",err)
		return
	}

	totalHandle, err := strconv.ParseInt(handleExported.String(), 10, 64)
	if err != nil {
		bylog.Error("calculateQPS totalHandle err=%s",err)
		return
	}

	if float64(totalConn)*totalTime != 0 {
		// take the average time per worker go-routine
		qps := float64(totalHandle) / (float64(totalConn) * (totalTime / float64(WorkerPoolInstance().Size())))
		qpsExported.Set(qps)
	}
}
