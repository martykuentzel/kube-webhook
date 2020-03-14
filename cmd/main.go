package main

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/MartyKuentzel/kube-webhook/pkg/mutate"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
}

func handleMutate(w http.ResponseWriter, r *http.Request) {

	log.Info("Start mutating ...")

	// read the body / request
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error("Cannot retrieve the body of the AdmissionReview request.%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// mutate the request
	mutated, err := mutate.Mutate(r.Context(), body)
	if err != nil {
		log.Errorf("Mutation failed.\n%v.", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// and write it back
	w.WriteHeader(http.StatusOK)
	w.Write(mutated)

	log.Debug("Mutation over.")
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", handleMutate)

	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}

	log.Infof("Listening on port: %s", s.Addr)
	log.Fatal(s.ListenAndServeTLS("./ssl/mutateme.pem", "./ssl/mutateme.key"))

}
