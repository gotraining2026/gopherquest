package main

import (
	"fmt"
	"os"
	"strings"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/api"
	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/cli"
)

func main() {
	apiURL := os.Getenv("GOPHERQUEST_API_URL")
	if strings.TrimSpace(apiURL) == "" {
		apiURL = "http://127.0.0.1:8080"
	}

	client, err := api.NewClient(apiURL)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	command := cli.NewRootCommand(client)
	command.SetOut(os.Stdout)
	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
