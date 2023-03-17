package common

import (
	"crypto/sha256"
	"fmt"
	"strings"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
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
