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
	fmt.Printf("This is kruiser watching namespace %v for gRPC services to publish\n", targetns)
	for {
		svcs, err := findServices(targetns, targetlabel)
		if err != nil {
			fmt.Errorf("Can't list services in namespace %v due to %v\n", targetns, err)
		}
		fmt.Printf("Found gRPC services %v\n", svcs)
		err = createProxies(svcs)
		if err != nil {
			fmt.Errorf("Can't create proxies due to %v\n", err)
		}
		time.Sleep(wdelay)
	}
}

func createProxies(svcs []string) error {
	type gRPCService struct {
		Name string
		Port string
	}
	s := gRPCService{"ping", "9000"}

	// create a location in NGINX conf per service
	// servert := `|
	// server {
	//   listen 8080 http2;

	//   access_log /dev/stdout;
	//   error_log /dev/stderr warn;

	//   _LOCATIONS_
	// }
	// `
	tmpl, err := template.New("service").Parse(`
	location /{{.Name}} {
        grpc_pass grpc://{{.Name}}:{{.Port}};
    }
    `)
	if err != nil {
		return err
	}
	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, s)
	if err != nil {
		return err
	}
	fmt.Println(buf.String())

	// create a ConfigMap nginxconf
	// kubectl "--namespace=kruiser" create configmap nginxconf --from-literal=config=TEMPLATE
	// re-deploy kruiser with new ConfigMap
	return nil
}

func findServices(namespace, label string) ([]string, error) {
	var res []string
	svcs, err := kubectl(true, "get",
		"--namespace="+namespace, "svc",
		"--selector="+label,
		"-o=custom-columns=:metadata.name",
		"--no-headers")
	if err != nil {
		return res, err
	}
	res = strings.Split(svcs, "\n")
	return res, nil
}
