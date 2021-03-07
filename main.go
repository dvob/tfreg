package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	var (
		gitlabMappingTemplate string
		tlsKey                string
		tlsCert               string
	)

	flag.StringVar(&tlsKey, "key", "tls.key", "TLS key file")
	flag.StringVar(&tlsCert, "cert", "tls.crt", "TLS cert file")
	flag.StringVar(&gitlabMappingTemplate, "template", "{{.Namespace}}/{{.Name}}", "TLS cert file")

	flag.Parse()

	gitlabModuleProvider, err := newGitlabModuleProvider(os.Getenv("GITLAB_TOKEN"), gitlabMappingTemplate)
	if err != nil {
		log.Fatal(err)
	}

	moduleHandler := newModuleHandler(gitlabModuleProvider)

	err = http.ListenAndServeTLS(":443", "tls.crt", "tls.key", logger(moduleHandler))
	if err != nil {
		log.Fatal(err)
	}
}
