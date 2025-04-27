package util

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lizc2003/vue-ssr-v8go/server/common/defs"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewGinEngine(env string) *gin.Engine {
	isProd := false
	if env == defs.EnvProd {
		isProd = true
		gin.SetMode(gin.ReleaseMode)
	} else if env == defs.EnvTest {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	e := gin.New()
	if isProd {
		gin.DisableConsoleColor()
		gin.DefaultWriter = &tlog.IoWriter{}
	}

	_ = e.SetTrustedProxies([]string{"127.0.0.0/8", "192.168.0.0/16", "172.16.0.0/12", "10.0.0.0/8"})
	return e
}

func GraceHttpServe(addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	var serveErr error
	serveEnd := make(chan struct{})

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			tlog.Error("http server start error:", err)
			serveErr = err
			close(serveEnd)
		}
	}()

	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-serveEnd:
		return serveErr
	case <-quit:
		tlog.Info("shutting down http server...")
		// The context is used to inform the server it has 5 seconds to finish
		// the request it is currently handling
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(timeoutCtx); err != nil {
			tlog.Error("http server shutdown error:", err)
		}
		return nil
	}
}
