package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/docker/distribution/configuration"
	dcontext "github.com/docker/distribution/context"
	"github.com/docker/distribution/registry/handlers"
	"github.com/docker/distribution/version"

	_ "github.com/docker/distribution/registry/storage/driver/azure"
)

func main() {
	handler, err := newDistributionHandler()
	if err != nil {
		log.Fatal(err)
	}
	registry := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}
	if err = registry.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// Run inside an Azure function.
func Run(r *http.Request) (resp *http.Response, err error) {

	handler, err := newDistributionHandler()
	if err != nil {
		return
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	resp = w.Result()
	return
}

func newDistributionHandler() (http.Handler, error) {
	dctx := dcontext.WithVersion(dcontext.Background(), version.Version)

	storageParameters, err := newParametersFromConnectionString(os.Getenv("AzureWebJobsStorage"))
	if err != nil {
		return nil, err
	}

	config := &configuration.Configuration{
		Version: "0.1",
		Storage: configuration.Storage{
			"azure": storageParameters,
		},
	}
	config.HTTP.Secret = os.Getenv("REGISTRY_HTTP_SECRET")

	app := handlers.NewApp(dctx, config)
	app.RegisterHealthChecks()
	return app, nil
}

// newParametersFromConnectionString creates a map[string]string from the connection string.
func newParametersFromConnectionString(input string) (map[string]interface{}, error) {
	parameters := map[string]interface{}{
		// use https by default
		"defaultendpointsprotocol": "https",
		"container":                "registry",
	}

	for _, pair := range strings.Split(input, ";") {
		if pair == "" {
			continue
		}

		equalDex := strings.IndexByte(pair, '=')
		if equalDex <= 0 {
			return nil, fmt.Errorf("Invalid connection segment %q", pair)
		}

		key := strings.ToLower(pair[:equalDex])
		value := pair[equalDex+1:]
		parameters[key] = value
	}

	return parameters, nil
}
