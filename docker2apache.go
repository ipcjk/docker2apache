package main

import (
	"context"
	"flag"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"net"
	"os"
	"strings"
	"text/template"
)

func main() {
	var ipTargetBind string

	domainToWatch := flag.String("d", "com.example", "reverse domain that is used for label scanning")
	labelValue := flag.String("v", "devtest.example.com", "label value we are looking for")
	portToSearch := uint16(*flag.Uint("p", 8009, "port we search, default 8009 for ajp"))
	proxyProto := flag.String("x", "ajp", "default protocol for the genereated proxy pass uris")

	ctx := context.TODO()

	flag.Parse()

	c, err := client.NewClientWithOpts()
	if err != nil {
		log.Fatal(err)
	}

	client.FromEnv(c)
	c.NegotiateAPIVersion(ctx)

	cs, err := c.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	var goodContainers []types.Container
	for _, container := range cs {
		if _, ok := container.Labels[*domainToWatch+".vhost"]; ok {
			if container.Labels[*domainToWatch+".vhost"] == *labelValue {
				for k, v := range container.Ports {
					if v.PrivatePort == portToSearch && container.State == "running" {
						container.Ports = append([]types.Port{}, container.Ports[k])
						goodContainers = append(goodContainers, container)
					}
				}
			}
		}
	}

	/* Was there $ENV-> DOCKER_HOST we needed to respect before?
	Then we guess, we need to rewrite the INADDR_ANY to the specific docker daemon ip
	=> not optimal but...
	*/
	if os.Getenv("DOCKER_HOST") != "" {
		url, err := client.ParseHostURL(os.Getenv("DOCKER_HOST"))
		if err != nil {
			log.Fatal("Cant parse DOCKER_HOST variable")
		}
		ipTargetBind, _, err = net.SplitHostPort(url.Host)
		if err != nil {
			log.Fatal("Cant parse DOCKER_HOST variable")
		}
	}

	if ipTargetBind == "" {
		ipTargetBind = "127.0.0.1"
	}

	if len(goodContainers) == 0 {
		os.Exit(1)
	}

	apacheTemplate := `
{{- range  $element := . -}}
{{$uri_maps_arr := Split (index .Labels "` + *domainToWatch + `.uri_maps") "," }}
{{- range $v, $uri_map := $uri_maps_arr -}}{{$front2back_uri := Split $uri_map ":" }}{{$port := index $element.Ports 0}}proxyPass {{ index $front2back_uri 0}} ` + *proxyProto + `://{{if eq $port.IP "0.0.0.0"}}` + ipTargetBind + `{{else}}{{$port.IP}}{{end}}:{{$port.PublicPort}}{{ index $front2back_uri 1}}
{{end -}}{{- end -}}`
	tplFuncMap := template.FuncMap{}
	tplFuncMap["Split"] = Split
	t, err := template.New("apache").Funcs(tplFuncMap).Parse(apacheTemplate)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(os.Stdout, goodContainers)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func Split(s, sep string) []string {
	return strings.Split(s, sep)
}
