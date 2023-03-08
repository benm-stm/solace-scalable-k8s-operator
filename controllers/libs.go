package controllers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Returns the resources Labels
func Labels(s *scalablev1alpha1.SolaceScalable) map[string]string {
	return map[string]string{
		"app": s.Name,
	}
}

// check if element o exist in the given slice
func IsItInSlice(o interface{}, list []string) bool {
	for _, b := range list {
		if b == o {
			return true
		}
	}
	return false
}

// Hash the given element with sha256
func AsSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

// Removes repeated int32 and 0 elements from given slice
func UniqueAndNonZero(intSlice []int32) []int32 {
	keys := make(map[int32]bool)
	list := []int32{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value && entry != 0 {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

/*
Calls the solace SEMPV2 Api
*/
func CallSolaceSempApi(
	u string,
	ctx context.Context,
	solaceAdminPassword string,
) (string, bool, error) {
	log := log.FromContext(ctx)
	client := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", false, err
	}
	req.SetBasicAuth("admin", solaceAdminPassword)
	resp, err := client.Do(req)
	if err != nil {
		return "", false, err
	}
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, err
	}
	if resp.StatusCode == 200 {
		return string(bodyText), true, nil
	} else {
		log.Info("solace Api call issue", u, resp.StatusCode)
		return "", false, nil
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
	ResValues := url.Values{}
	for key, value := range v {
		ResValues.Add(key, value)
	}
	resUrl.RawQuery = ReformatForSolace(ResValues.Encode())
	return resUrl.String()
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

/*
RFC 1035 Label Names
Some resource types require their names to follow the DNS label standard as defined in RFC 1035. This means the name must:

	1- contain at most 63 characters
	2- contain only lowercase alphanumeric characters or '-'
	3- start with an alphabetic character
	4- end with an alphanumeric character
*/
func SanityzeForSvcName(s string) string {
	forbiddenChars := []string{"#", "\\", "/", ")", "(", ":"}
	if len(s) <= 63 {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, " ", "-")
		s = strings.ReplaceAll(s, "_", "-")
		for _, v := range forbiddenChars {
			s = strings.ReplaceAll(s, v, "")
		}
	}
	return s
}

/*
Searches for the next available int32 based on a given array of int32, see below's example
  - ap = [1025, 1026, 1028]
  - p = 1026
  - return 1027
*/
func NextAvailablePort(
	ap []int32,
	p int32,
) int32 {
	for i := range ap {
		if p == ap[i] {
			return NextAvailablePort(
				ap,
				p+1,
			)
		}
	}
	return p
}

// Solace doesn't accept "," encoding
func ReformatForSolace(s string) string {
	return strings.Replace(s, "%2C", ",", -1)
}
