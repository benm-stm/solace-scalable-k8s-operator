package controllers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func Labels(s *scalablev1alpha1.SolaceScalable) map[string]string {
	// Fetches and sets labels
	return map[string]string{
		"app": s.Name,
	}
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func AsSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func Unique(intSlice []int32) []int32 {
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

func GetEnv(s *scalablev1alpha1.SolaceScalable, selector string) string {
	for e := 0; e < int(len(s.Spec.Container.Env)); e++ {
		if s.Spec.Container.Env[e].Name == selector {
			return s.Spec.Container.Env[e].Value
		}
	}
	return ""
}

func CleanJsonResponse(s string, r string) []int32 {
	var ret []int32
	re, _ := regexp.Compile(r)
	submatchall := re.FindAllStringSubmatch(s, -1)
	for _, s := range submatchall {
		x, err := strconv.ParseInt(s[1], 10, 32)
		if err != nil {
			panic(err)
		} else if x != 0 {
			ret = append(ret, int32(x))
		}
	}
	return ret
}

func CallSolaceSempApi(s *scalablev1alpha1.SolaceScalable, apiPath string, ctx context.Context) (string, bool, error) {
	log := log.FromContext(ctx)
	var retErr error
	for i := 0; i < int(s.Spec.Replicas); i++ {
		url := "n" + strconv.Itoa(i) + "." + s.Spec.ClusterUrl

		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://"+url+"/SEMP/v2"+apiPath, nil)
		if err != nil {
			retErr = err
		}
		req.SetBasicAuth("admin", GetEnv(s, "username_admin_password"))
		resp, err := client.Do(req)
		if err != nil {
			retErr = err
		}
		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			retErr = err
		}
		if resp.StatusCode == 200 {
			return string(bodyText), true, nil
		} else {
			log.Info("solace Url unreachable " + url)
		}
	}
	log.Error(retErr, "All solace Urls are unreachable ")
	return "", false, retErr
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

func EnvVars(s *scalablev1alpha1.SolaceScalableSpec) []corev1.EnvVar {
	var env []corev1.EnvVar

	for _, s := range s.Container.Env {
		env = append(env, corev1.EnvVar{Name: s.Name, Value: s.Value})
	}
	return env
}