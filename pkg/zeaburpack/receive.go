package zeaburpack

import (
	"context"
	"net/http"
	"time"
)

func receiveFiles(addr string) {

	shouldEnd := false

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc(
		"/end", func(w http.ResponseWriter, r *http.Request) {
			shouldEnd = true
		},
	)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
			shouldEnd = true
		}
	}()

	for !shouldEnd {
		time.Sleep(1 * time.Second)
	}

	err := server.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
}
