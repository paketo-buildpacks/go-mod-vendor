package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

func main() {

	http.HandleFunc("/", callHelper)
	fmt.Println("listening...")

	port := "8080"
	if systemPort := os.Getenv("PORT"); systemPort != "" {
		port = systemPort
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

func callHelper(res http.ResponseWriter, req *http.Request) {
	cmd := exec.Command("helper")
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}

	fmt.Fprintf(res, string(output))
}
