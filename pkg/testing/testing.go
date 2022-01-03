// Copyright 2022 Thiago Souza <tcostasouza@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func StartEosio(t *testing.T) (int, int) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}
	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "tsouza/eosio-dev",
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

	return shipPort, testRpcPort
}

func StopEosio(t *testing.T) {
	err := testPool.Purge(testRes)
	if err != nil {
		t.Fatal(err)
	}
	testPool = nil
	testRes = nil
}