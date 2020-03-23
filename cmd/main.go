package main

import (
	"context"
	"net/http"
	"time"

	"github.com/MartyKuentzel/kube-webhook/pkg/mutate"
	"github.com/MartyKuentzel/kube-webhook/pkg/vault"
	log "github.com/sirupsen/logrus"
)

func main() {

	// TODO introduce Flags
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{})

	// TODO: ctx recherieren
	ctx := context.Background()
	secHook := &mutate.SecHook{Vault: vault.New(ctx)}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", secHook.HandleMutate)

	//TODO: best practice http Server (go fucntions)
	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}

	log.Infof("Listening on port: %s", s.Addr)
	log.Fatal(s.ListenAndServeTLS("./ssl/kube-secHook.pem", "./ssl/kube-secHook.key"))
}
