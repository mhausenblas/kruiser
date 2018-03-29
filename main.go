package main

import (
	"fmt"
	"os"
	"time"
)

var (
	debug = false
	// if KRUISER_TARGET_NAMESPACE not set, watch default namespace:
	targetns = "default"
	// if KRUISER_TARGET_LABEL not set, watch for label grpc=expose:
	targetlabel = "grpc=expose"
	// checking for gRPC services every 5 sec:
	wdelay = time.Duration(5) * time.Second
)

func init() {
	if d := os.Getenv("KRUISER_DEBUG"); d != "" {
		debug = true
	}
	if tns := os.Getenv("KRUISER_TARGET_NAMESPACE"); tns != "" {
		targetns = tns
	}
	if tl := os.Getenv("KRUISER_TARGET_LABEL"); tl != "" {
		targetlabel = tl
	}
}

func main() {
	fmt.Printf("This is kruiser watching namespace %v for gRPC services to publish\n", targetns)
	for {
		svcs, err := findSvc(targetns, targetlabel)
		if err != nil {
			fmt.Errorf("Can't list services in namespace %v", targetns)
		}
		fmt.Println(svcs)
		time.Sleep(wdelay)
	}
}

func findSvc(namespace, label string) (string, error) {
	svcs, err := kubectl(true, "get",
		"--namespace="+namespace, "svc",
		"--selector="+label,
		"-o=custom-columns=:metadata.name",
		"--no-headers")
	if err != nil {
		return svcs, err
	}
	return svcs, nil
}
