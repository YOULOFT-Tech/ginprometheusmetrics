package main

import (
    "github.com/gin-gonic/gin"
    //"gitlab.ylwnl.com/middleware/ginprometheusmetrics"
    "log"
    "net/http"
)

func main() {
    engine := gin.Default()
    setPrometheus(engine)

    engine.GET("/test", func(context *gin.Context) {
        context.Writer.WriteString("hello world")
    })
    engine.Run(":9090")
    if err := http.ListenAndServe(":9090", engine); err != nil {
        log.Fatal(err)
    }
}

func setPrometheus(engine *gin.Engine) {
    //namespace := "test-szl"
    //hostName, _ := os.Hostname()
    //opts := ginprometheusmetrics.PrometheusOpts{
    //	PushInterval:     uint8(30),
    //	PushGateWayUrl:   "http://127.0.0.1:8080",
    //	JobName:          "test-szl",
    //	Instance:         hostName,         //pod-name or hostname
    //	MonitorUri:       []string{},       //empty slice monitor all uri.
    //	ExcludeMethod:    []string{"HEAD"}, //exclude http method
    //	Percentage:       90,
    //	ExcludeURLPrefix: []string{"/test", "favicon.ico"},
    //}
    //
    //ginprometheusmetrics.NewPrometheus(namespace, opts, nil).Use(engine)

}
