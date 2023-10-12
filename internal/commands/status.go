package commands

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
)

// StatusCommand configures the command name, flags, and action.
var StatusCommand = cli.Command{
	Name:   "status",
	Usage:  "Checks if the Web server is running",
	Action: statusAction,
}

// statusAction checks if the Web server is running.
func statusAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)

	// Create new http.Client instance.
	//
	// NOTE: Timeout specifies a time limit for requests made by
	// this Client. The timeout includes connection time, any
	// redirects, and reading the response body. The timer remains
	// running after Get, Head, Post, or Do return and will
	// interrupt reading of the Response.Body.
	client := &http.Client{Timeout: 10 * time.Second}

	// Connect to unix socket?
	if unixSocket := conf.HttpSocket(); unixSocket != "" {
		client.Transport = &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", unixSocket)
			},
		}
	}

	url := fmt.Sprintf("http://%s:%d/api/v1/status", conf.HttpHost(), conf.HttpPort())

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return err
	}

	var status string

	if resp, err := client.Do(req); err != nil {
		return fmt.Errorf("cannot connect to %s:%d", conf.HttpHost(), conf.HttpPort())
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("server running at %s:%d, bad status %d\n", conf.HttpHost(), conf.HttpPort(), resp.StatusCode)
	} else if body, err := io.ReadAll(resp.Body); err != nil {
		return err
	} else {
		status = string(body)
	}

	message := gjson.Get(status, "status").String()

	if message != "" {
		fmt.Println(message)
	} else {
		fmt.Println("unknown")
	}

	return nil
}
