package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	if os.Getenv("CONFIG_SERVER_PORT") == "" {
		fmt.Fprintln(os.Stderr, "Error: missing $CONFIG_SERVER_PORT")
		os.Exit(1)
	}

	// Emulate an external configuration service
	http.HandleFunc("/config/", config)
	fmt.Println("listening 0.0.0.0:" + os.Getenv("CONFIG_SERVER_PORT") + "...")
	err := http.ListenAndServe(":"+os.Getenv("CONFIG_SERVER_PORT"), nil)
	if err != nil {
		panic(err)
	}
}

type Config struct {
	Scope    string
	Password string
}

func config(res http.ResponseWriter, req *http.Request) {
	config := Config{"some-service.admin", "not-a-real-p4$$w0rd"}

	js, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}

	fmt.Println("Received a request for config.")
	fmt.Fprintln(res, string(js))
}
