package debug

import (
	"net/http"
	"net/http/pprof"
	"runtime"

	"github.com/gin-gonic/gin"
)

func RegisterPprof(r *gin.Engine) {
	debug := r.Group("/debug/pprof")
	{
		debug.GET("/", gin.WrapF(pprof.Index))
		debug.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		debug.GET("/profile", gin.WrapF(pprof.Profile))
		debug.GET("/symbol", gin.WrapF(pprof.Symbol))
		debug.POST("/symbol", gin.WrapF(pprof.Symbol))
		debug.GET("/trace", gin.WrapF(pprof.Trace))

		debug.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		debug.GET("/block", gin.WrapH(pprof.Handler("block")))
		debug.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		debug.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		debug.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		debug.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}
}

func RegisterHealthz(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		c.JSON(http.StatusOK, gin.H{
			"status":         "ok",
			"goroutines":     runtime.NumGoroutine(),
			"heap_alloc_mb":  float64(m.HeapAlloc) / 1024 / 1024,
			"heap_objects":   m.HeapObjects,
			"sys_mb":         float64(m.Sys) / 1024 / 1024,
			"gc_cycles":      m.NumGC,
			"gc_pause_total": m.PauseTotalNs,
		})
	})
}
