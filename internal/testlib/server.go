// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package testlib

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"

	"github.com/itential/ipctl/internal/profile"
	"github.com/itential/ipctl/pkg/client"
)

var (
	mux       *http.ServeMux
	server    *httptest.Server
	iapclient client.Client
)

func Setup() client.Client {
	mux = http.NewServeMux()

	server = httptest.NewServer(mux)

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("TEST_TOKEN")
	})

	u, err := url.Parse(server.URL)

	host, strPort, err := net.SplitHostPort(u.Host)
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(strPort)
	if err != nil {
		panic(err)
	}

	prof := &profile.Profile{
		Host:   host,
		Port:   port,
		UseTLS: u.Scheme == "https",
	}

	iapclient = client.New(context.TODO(), prof)

	return iapclient
}

func Teardown() {
	server.Close()
}

func Fixture(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("!!! WARNING: %s", msg)
}

func addResponse(uri, body, method string, statusCode int) {
	mux.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != method {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(statusCode)

		if body != "" {
			fmt.Fprint(w, body)
		}
	})
}

func AddGetResponseToMux(uri, body string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	if statusCode != http.StatusOK {
		warn("non standard status code return for GET %s, expected %v, want %v\n", uri, http.StatusOK, statusCode)
	}
	addResponse(uri, body, http.MethodGet, statusCode)
}

func AddGetErrorToMux(uri, body string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	addResponse(uri, body, http.MethodGet, statusCode)
}

func AddPostResponseToMux(uri, body string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusCreated
	}
	if statusCode != http.StatusCreated {
		warn("non standard status code return for POST %s, expected %v, want %v\n", uri, http.StatusCreated, statusCode)
	}
	addResponse(uri, body, http.MethodPost, statusCode)
}

func AddPostErrorToMux(uri, body string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	addResponse(uri, body, http.MethodPost, statusCode)
}

func AddDeleteResponseToMux(uri, body string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	if statusCode != http.StatusOK {
		warn("non standard status code return for DELETE %s, expected %v, want %v\n", uri, http.StatusOK, statusCode)
	}

	addResponse(uri, body, http.MethodDelete, statusCode)
}

func AddDeleteErrorToMux(uri, body string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	addResponse(uri, body, http.MethodDelete, statusCode)
}

func AddPatchResponseToMux(uri, body string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	if statusCode != http.StatusOK {
		warn("non standard status code return for PATCH %s, expected %v, want %v\n", uri, http.StatusOK, statusCode)
	}
	addResponse(uri, body, http.MethodPatch, statusCode)
}

func AddPatchErrorToMux(uri, body string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	addResponse(uri, body, http.MethodPatch, statusCode)
}

func AddPutResponseToMux(uri, body string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	if statusCode != http.StatusOK {
		warn("non standard status code return for PUT %s, expected %v, want %v\n", uri, http.StatusOK, statusCode)
	}
	addResponse(uri, body, http.MethodPut, statusCode)
}

func AddPutErrorToMux(uri, body string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	addResponse(uri, body, http.MethodPut, statusCode)
}

// AddHandlerToMux registers a custom handler for the given URI pattern,
// allowing tests to inspect the incoming request directly.
func AddHandlerToMux(uri string, handler http.HandlerFunc) {
	mux.HandleFunc(uri, handler)
}
