package mutate

import (
	"io/ioutil"
	"net/http"

	"github.com/MartyKuentzel/kube-webhook/pkg/vault"
	log "github.com/sirupsen/logrus"
)

// SecHook initliazes Vault Client and handles Mutation
type SecHook struct {
	Vault vault.VaultClient
}

//HandleMutate takes care of request by reading body in passing it to the Mutate Function
func (secHook *SecHook) HandleMutate(w http.ResponseWriter, r *http.Request) {

	log.Info("Start mutating ...")

	//TODO: Use this log to create a working mutation in secHook test
	log.Info("REQUEST__:", r)

	// read the body / request
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Errorf("Cannot retrieve the body of the AdmissionReview request.%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// mutate the request with vault client
	mutated, err := mutate(r.Context(), body, secHook.Vault)
	if err != nil {
		log.Errorf("Mutation failed.\n%v.", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// and write it back
	w.WriteHeader(http.StatusOK)
	w.Write(mutated)

	log.Debug("...Mutation over.")
}
