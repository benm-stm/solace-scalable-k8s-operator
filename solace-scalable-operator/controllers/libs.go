package controllers

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	scalablev1alpha1 "solace.io/api/v1alpha1"
)

func labels(s *scalablev1alpha1.SolaceScalable) map[string]string {
	// Fetches and sets labels
	return map[string]string{
		"app": s.Name,
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func asSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func unique(intSlice []int32) []int32 {
	keys := make(map[int32]bool)
	list := []int32{}
	for _, entry := range intSlice {
		//if _, value := keys[entry]; !value {
		// remove 0
		if _, value := keys[entry]; !value && entry == 0 {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func getEnv(s *scalablev1alpha1.SolaceScalable, selector string) string {
	for e := 0; e < int(len(s.Spec.Container.Env)); e++ {
		if s.Spec.Container.Env[e].Name == selector {
			return s.Spec.Container.Env[e].Value
		}
	}
	return ""
}

func cleanJsonResponse(s string, r string) []int32 {
	var ret []int32
	re, _ := regexp.Compile(r)
	submatchall := re.FindAllStringSubmatch(s, -1)
	for i := 0; i < len(submatchall); i++ {
		x, err := strconv.ParseInt(submatchall[i][1], 10, 32)
		if err != nil {
			panic(err)
		} else if x != 0 {
			ret = append(ret, int32(x))
		}
	}
	return ret
}

func callSolaceSempApi(s *scalablev1alpha1.SolaceScalable, apiPath string) string {
	for i := 0; i < int(s.Spec.Replicas); i++ {
		//name := s.Name + "-" + strconv.Itoa(i)
		//url := s.name + "." + s.Namespace + ".svc.cluster.local"
		//url := "35.195.99.220"
		url := "n" + strconv.Itoa(i) + "." + s.Spec.ClusterUrl

		client := &http.Client{}
		//req, err := http.NewRequest("GET", "http://"+url+":8080/SEMP/v2"+apiPath, nil)
		req, err := http.NewRequest("GET", "http://"+url+"/SEMP/v2"+apiPath, nil)
		req.SetBasicAuth("admin", getEnv(s, "username_admin_password"))
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		bodyText, err := ioutil.ReadAll(resp.Body)
		return string(bodyText)
	}
	return ""
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

/*
func portsRange(s string) []corev1.ContainerPort {
	var ports []corev1.ContainerPort
	// append default ones
	ports = append(ports, corev1.ContainerPort{ContainerPort: 8080})
	ports = append(ports, corev1.ContainerPort{ContainerPort: 55555})
	if len(s) > 0 {
		split := strings.Split(s, "-")
		portOne, err := strconv.ParseInt(split[0], 10, 32)
		if err != nil {
			panic(err)
		}
		portTwo, err := strconv.ParseInt(split[1], 10, 32)
		if err != nil {
			panic(err)
		}
		for i := portOne; i <= portTwo; i++ {
			ports = append(ports, corev1.ContainerPort{ContainerPort: int32(i)})
		}
	}
	return ports
}*/

func envVars(s *scalablev1alpha1.SolaceScalableSpec) []corev1.EnvVar {
	var env []corev1.EnvVar

	for i := 0; i < len(s.Container.Env); i++ {
		env = append(env, corev1.EnvVar{Name: s.Container.Env[i].Name, Value: s.Container.Env[i].Value})
	}
	return env
}
