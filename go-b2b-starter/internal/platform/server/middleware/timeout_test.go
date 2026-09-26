package middleware

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/gin-gonic/gin"
)

func TestTimeoutPreservesResponseAndReleasesContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		type requestKey struct{}
		parent := context.WithValue(t.Context(), requestKey{}, "request-value")
		var handlerContext context.Context
		router := gin.New()
		router.Use(Timeout(5 * time.Second))
		router.GET("/", func(c *gin.Context) {
			handlerContext = c.Request.Context()
			if handlerContext.Value(requestKey{}) != "request-value" {
				t.Fatal("request context values were lost")
			}
			if deadline, ok := handlerContext.Deadline(); !ok || deadline.Sub(time.Now()) != 5*time.Second {
				t.Fatalf("unexpected request deadline: %v, present=%v", deadline, ok)
			}
			c.Header("X-Result", "preserved")
			c.String(http.StatusCreated, "created")
		})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequestWithContext(parent, http.MethodGet, "/", nil))
		if response.Code != http.StatusCreated || response.Body.String() != "created" || response.Header().Get("X-Result") != "preserved" {
			t.Fatalf("handler response changed: %d %q %v", response.Code, response.Body.String(), response.Header())
		}
		if handlerContext.Err() != context.Canceled || parent.Err() != nil {
			t.Fatalf("request context cleanup: child=%v parent=%v", handlerContext.Err(), parent.Err())
		}
	})
}

func TestTimeoutBoundsCooperativeWork(t *testing.T) {
	for _, tc := range []struct {
		name   string
		parent time.Duration
		want   time.Duration
	}{
		{name: "request deadline", parent: time.Hour, want: 5 * time.Second},
		{name: "earlier parent deadline", parent: time.Second, want: time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				parent, cancel := context.WithTimeout(t.Context(), tc.parent)
				defer cancel()
				completedWork := false
				router := gin.New()
				router.Use(Timeout(5 * time.Second))
				router.GET("/", func(c *gin.Context) {
					operation := time.NewTimer(time.Hour)
					defer operation.Stop()
					select {
					case <-operation.C:
						completedWork = true
						c.String(http.StatusOK, "late result")
					case <-c.Request.Context().Done():
						if c.Request.Context().Err() != context.DeadlineExceeded {
							t.Errorf("unexpected cancellation: %v", c.Request.Context().Err())
						}
						// Error translation belongs to the handler, which is still the
						// only owner of this response after cooperative work stops.
						c.String(http.StatusGatewayTimeout, "operation timed out")
					}
				})
				response := httptest.NewRecorder()
				start := time.Now()
				router.ServeHTTP(response, httptest.NewRequestWithContext(parent, http.MethodGet, "/", nil))
				if elapsed := time.Since(start); elapsed != tc.want {
					t.Fatalf("cooperative handler finished after %v, want %v", elapsed, tc.want)
				}
				if completedWork || response.Code != http.StatusGatewayTimeout || response.Body.String() != "operation timed out" {
					t.Fatalf("late work or replaced response: completed=%v status=%d body=%q", completedWork, response.Code, response.Body.String())
				}
				synctest.Wait()
				if response.Body.String() != "operation timed out" {
					t.Fatal("response changed after the handler returned")
				}
			})
		})
	}
}

func TestTimeoutPropagatesParentCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		parent, cancel := context.WithCancel(t.Context())
		defer cancel()
		entered := make(chan struct{})
		go func() {
			<-entered
			cancel()
		}()
		router := gin.New()
		router.Use(Timeout(time.Hour))
		router.GET("/", func(c *gin.Context) {
			close(entered)
			<-c.Request.Context().Done()
			if c.Request.Context().Err() != context.Canceled {
				t.Errorf("parent cancellation lost: %v", c.Request.Context().Err())
			}
			c.Status(http.StatusNoContent)
		})
		response := httptest.NewRecorder()
		start := time.Now()
		router.ServeHTTP(response, httptest.NewRequestWithContext(parent, http.MethodGet, "/", nil))
		if time.Since(start) != 0 || response.Code != http.StatusNoContent || response.Body.Len() != 0 {
			t.Fatalf("canceled request waited or gained a response body: %d %q", response.Code, response.Body.String())
		}
	})
}

func TestTimeoutDoesNotOverwritePartialResponse(t *testing.T) {
	for _, stream := range []bool{false, true} {
		name := "partial"
		if stream {
			name = "flushed stream"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				router := gin.New()
				router.Use(Timeout(time.Second))
				router.GET("/", func(c *gin.Context) {
					c.String(http.StatusAccepted, "first chunk\n")
					if stream {
						c.Writer.Flush()
					}
					select {
					case <-time.After(time.Hour):
						c.String(http.StatusAccepted, "late chunk\n")
					case <-c.Request.Context().Done():
						return // A started response must not receive a second error document.
					}
				})
				response := httptest.NewRecorder()
				router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
				synctest.Wait()
				if response.Code != http.StatusAccepted || response.Body.String() != "first chunk\n" || response.Flushed != stream {
					t.Fatalf("partial response corrupted: %d %q flushed=%v", response.Code, response.Body.String(), response.Flushed)
				}
			})
		})
	}
}

func TestTimeoutLeavesPanicRecoveryOnRequestGoroutine(t *testing.T) {
	var handlerContext context.Context
	var recovered any
	router := gin.New()
	router.Use(gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, value any) {
		recovered = value
		if handlerContext.Err() != context.Canceled {
			t.Error("timeout context was not released while unwinding the handler")
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	router.Use(Timeout(time.Hour))
	router.GET("/", func(c *gin.Context) {
		handlerContext = c.Request.Context()
		panic("handler panic")
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if recovered != "handler panic" || response.Code != http.StatusInternalServerError || response.Body.Len() != 0 {
		t.Fatalf("panic escaped request recovery: recovered=%v status=%d body=%q", recovered, response.Code, response.Body.String())
	}
}

func TestTimeoutCannotInterruptNonCooperativeHandler(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		returned := make(chan struct{})
		released := false
		defer func() {
			if !released {
				close(release)
			}
		}()
		var handlerContext context.Context
		router := gin.New()
		router.Use(Timeout(time.Second))
		router.GET("/", func(c *gin.Context) {
			handlerContext = c.Request.Context()
			<-release // Deliberately ignores context to establish the safety limit.
			c.String(http.StatusOK, "handler still owns response")
		})
		response := httptest.NewRecorder()
		go func() {
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
			close(returned)
		}()
		synctest.Wait()
		time.Sleep(2 * time.Second)
		synctest.Wait()
		if handlerContext.Err() != context.DeadlineExceeded {
			t.Fatalf("request deadline did not expire: %v", handlerContext.Err())
		}
		select {
		case <-returned:
			t.Fatal("middleware returned while its handler could still use the Gin context")
		default:
		}
		if response.Body.Len() != 0 {
			t.Fatalf("middleware wrote a competing timeout response: %q", response.Body.String())
		}
		close(release)
		released = true
		<-returned
		if response.Code != http.StatusOK || response.Body.String() != "handler still owns response" {
			t.Fatalf("handler lost writer ownership: %d %q", response.Code, response.Body.String())
		}
	})
}

func TestTimeoutObservesHTTPClientDisconnect(t *testing.T) {
	entered := make(chan struct{})
	stopped := make(chan error, 1)
	cleanup := make(chan struct{})
	router := gin.New()
	router.Use(Timeout(time.Hour))
	router.GET("/", func(c *gin.Context) {
		close(entered)
		select {
		case <-c.Request.Context().Done():
			stopped <- c.Request.Context().Err()
		case <-cleanup:
			return
		}
	})
	server := httptest.NewServer(router)
	defer server.Close()
	defer close(cleanup) // Also release the test server if a cancellation assertion fails.
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := server.Client()
	client.Timeout = 5 * time.Second
	clientDone := make(chan error, 1)
	go func() {
		response, err := client.Do(request)
		if response != nil {
			response.Body.Close()
		}
		clientDone <- err
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("HTTP request did not reach the handler")
	}
	cancel()
	select {
	case err := <-stopped:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("disconnect error = %v, want context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("client disconnect did not cancel handler work")
	}
	if err := <-clientDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("client request error = %v, want context.Canceled", err)
	}
}

func TestTimeoutLeavesNetworkWriteDeadlineToHTTPServer(t *testing.T) {
	type outcome struct{ contextError, writeError error }
	finished := make(chan outcome, 1)
	cleanup := make(chan struct{})
	router := gin.New()
	// A shorter network deadline deliberately expires first. It does not itself
	// interrupt handler work or cancel the request context while no I/O happens.
	router.Use(Timeout(100 * time.Millisecond))
	router.GET("/", func(c *gin.Context) {
		select {
		case <-c.Request.Context().Done():
		case <-cleanup:
			return
		}
		contextError := c.Request.Context().Err()
		// Exceed net/http's response buffer to observe the actual socket error.
		_, writeError := c.Writer.Write(bytes.Repeat([]byte("x"), 64*1024))
		finished <- outcome{contextError, writeError}
	})
	server := httptest.NewUnstartedServer(router)
	server.Config.WriteTimeout = 20 * time.Millisecond
	server.Start()
	defer server.Close()
	defer close(cleanup)
	client := server.Client()
	client.Timeout = 5 * time.Second
	response, err := client.Get(server.URL)
	if response != nil {
		response.Body.Close()
	}
	if err == nil {
		t.Fatal("client received a response after the network write deadline")
	}
	select {
	case result := <-finished:
		if !errors.Is(result.contextError, context.DeadlineExceeded) {
			t.Fatalf("handler ended before its cooperative deadline: %v", result.contextError)
		}
		var timeoutError net.Error
		if !errors.As(result.writeError, &timeoutError) || !timeoutError.Timeout() {
			t.Fatalf("late network write error = %v, want timeout", result.writeError)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not finish after its cooperative deadline")
	}
}
