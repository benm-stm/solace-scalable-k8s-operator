package handler

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type SolaceValues struct {
	*url.Values
}

/*
Calls the solace SEMPV2 Api
*/
func CallSolaceSempApi(
	u string,
	ctx context.Context,
	solaceAdminPassword string,
) ([]byte, bool, error) {
	log := log.FromContext(ctx)
	client := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, false, err
	}
	req.SetBasicAuth("admin", solaceAdminPassword)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}
	if resp.StatusCode == 200 {
		return body, true, nil
	} else {
		log.Info("solace Api call issue", u, resp.StatusCode)
		return nil, false, nil
	}
}

/*
Construct Solace Semp URL from given parameters
  - url = scalable.solace.io
  - nodeNumber = 0
  - params map[string]string{"p1": "p1", "p2": "p2"}
  - return http://scalable.solace.io/SEMP/v2?p1=p1&p2=p2
*/
func ConstructSempUrl(
	s scalablev1alpha1.SolaceScalable,
	n int,
	p string,
	v map[string]string,
) string {
	var host string
	if s.Spec.ClusterUrl == "" {
		// use console svc kube dns
		// Can't be tested with `make run`, the operator should be installed in the kube itself
		port := "8080"
		host = s.ObjectMeta.Name + "-" + strconv.Itoa(n) + "." + s.ObjectMeta.Namespace + ":" + port
	} else {
		// use dns or ip exposed in the net
		host = "n" + strconv.Itoa(n) + "." + s.Spec.ClusterUrl
	}

	resUrl := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   "/SEMP/v2" + p,
	}
	ResValues := &SolaceValues{&url.Values{}}
	for key, value := range v {
		ResValues.Add(key, value)
	}
	resUrl.RawQuery = ResValues.EncodeForSolace()
	return resUrl.String()
}

// Solace doesn't accept "," encoding
func (s *SolaceValues) EncodeForSolace() string {
	return strings.Replace(s.Encode(), "%2C", ",", -1)
}
