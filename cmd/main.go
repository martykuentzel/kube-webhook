package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/MartyKuentzel/kube-webhook/pkg/mutate
	log "github.com/sirupsen/logrus"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handleroot was called \n")
	fmt.Fprintf(w, "hello %q", html.EscapeString(r.URL.Path))
}

func handleMutate(w http.ResponseWriter, r *http.Request) {

	log.Debugf("Start mutating ...")

	// read the body / request
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Errorf("Cannot retrieve the body of the AdmissionReview request.%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	// mutate the request
	mutated, err := mutate.Mutate(r.Context(), body)
	if err != nil {
		log.Errorf("Mutation failed.\n%v.", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	// and write it back
	w.WriteHeader(http.StatusOK)
	w.Write(mutated)

	log.Debugf("Mutation over.")

}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/mutate", handleMutate)

	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}

	log.Printf("Listening on port: %s\n", s.Addr)
	log.Fatal(s.ListenAndServeTLS("./ssl/mutateme.pem", "./ssl/mutateme.key"))

}
