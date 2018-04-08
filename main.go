package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"
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
	fmt.Printf("This is kruiser watching namespace %v for deployments labelled with %v so that I can publish services to proxy gRPC traffic\n", targetns, targetlabel)
	for {
		deploys, err := find(targetns, targetlabel)
		if err != nil {
			fmt.Errorf("Can't list deployments in namespace %v due to %v\n", targetns, err)
		}
		switch {
		case deploys[0] == "":
			fmt.Printf("Didn't find any deployments to proxy\n")
		default:
			fmt.Printf("Found deployments %v to create gRPC proxies for\n", deploys)
			err = proxy(targetns, deploys)
			if err != nil {
				fmt.Errorf("Can't create gRPC proxies due to %v\n", err)
			}
		}
		time.Sleep(wdelay)
	}
}

// proxy takes a list of deployment names
// and creates an Ambassador-backed service
// for each that proxies traffic to its pods.
func proxy(namespace string, deploys []string) error {
	// 1. create proxy services
	type gRPCService struct {
		Name          string
		FQServiceName string
		NodePort      string
		Port          string
		TargetPort    string
	}
	svcs := bytes.NewBufferString("")

	for _, deploy := range deploys {
		s := gRPCService{
			deploy,
			"yages.Echo",
			"30000",
			"9000",
			"9000",
		}
		tmpl, err := template.New("service").Parse(proxy_template)
		if err != nil {
			return err
		}
		err = tmpl.Execute(svcs, s)
		if err != nil {
			return err
		}
	}
	// 2. write out to tpm file:
	fmt.Printf("%v", svcs.String())

	// 3. apply tmp file containing service proxies:
	// res, err := kubectl(true, "apply",
	// 	"--namespace="+namespace,
	// 	"-f="+tmpsvcproxies)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("%v", res)
	return nil
}

// find queries the given Kubernetes namespace
// for deployments with the given label and
// returns a list of matching deployment names.
func find(namespace, label string) ([]string, error) {
	var res []string
	deploys, err := kubectl(true, "get",
		"--namespace="+namespace, "deploy",
		"--selector="+label,
		"-o=custom-columns=:metadata.name",
		"--no-headers")
	if err != nil {
		return res, err
	}
	// fmt.Printf("RAW: [%v] \n", deploys)
	res = strings.Split(deploys, "\n")
	return res, nil
}

// getconf queries the annotation of a deployment
// to get nodePort, port, targetPort, and the
/// fully qualified service name in the form package.Service
// func getconf(deploy string) (string, error) {

// }
