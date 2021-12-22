package testing

import (
	"bufio"
	"fmt"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	"io"
	"net/http"
	"strconv"
	"testing"
)

var (
	testRpcPort int
	testPool	*dockertest.Pool
	testRes		*dockertest.Resource
)

func StartEosio(t *testing.T) int {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}
	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "qryio/eosio-dev",
		Tag:          "latest",
		PortBindings: map[dc.Port][]dc.PortBinding{
			"8080/tcp": {{ HostIP: "0.0.0.0", HostPort: "8080/tcp" }},
			"8888/tcp": {{ HostIP: "0.0.0.0", HostPort: "8888/tcp" }},
		},
	}, func(config *dc.HostConfig) {
		config.AutoRemove = true
	})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	if err := pool.Retry(func() error {
		_, err = http.Get(fmt.Sprint("http://localhost:", res.GetPort("8888/tcp"), "/"))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		_ = pool.Purge(res)
		t.Fatal(err)
	}

	booting := true
	r, w := io.Pipe()
	go func() {
		for booting {
			if err := pool.Client.Logs(dc.LogsOptions{
				Container: res.Container.ID,
				RawTerminal:  false,
				Stdout: true,
				OutputStream: w,
				Stderr: true,
				ErrorStream: w,
			}); err != nil {
				_ = pool.Purge(res)
				t.Fatalf("Error reading eosio node logs: %s", err)
			}
		}
		_ = w.Close()
		_ = r.Close()
	}()

	s := bufio.NewScanner(r)
	for s.Scan() {
		l := s.Text()
		if l == "EOSIO_BOOT_COMPLETE" {
			booting = false
			break
		}
	}
	if s.Err() != nil {
		_ = pool.Purge(res)
		t.Fatalf("Error reading eosio node logs: %s", err)
	}

	shipPort, _ := strconv.Atoi(res.GetPort("8080/tcp"))
	testRpcPort, _ = strconv.Atoi(res.GetPort("8888/tcp"))
	testPool = pool
	testRes = res

	return shipPort
}

func StopEosio(t *testing.T) {
	err := testPool.Purge(testRes)
	if err != nil {
		t.Fatal(err)
	}
	testPool = nil
	testRes = nil
}