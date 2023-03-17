package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/benm-stm/solace-scalable-k8s-operator/tests"
)

func TestCallSolaceSempApi(t *testing.T) {

	mux := http.NewServeMux()
	mux.HandleFunc("/SEMP/v2", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"success": true}`))
		if err != nil {
			print(err)
		}
	})

	testTable := []struct {
		name             string
		server           *httptest.Server
		expectedResponse []byte
		expectedBool     bool
		expectedErr      error
	}{
		{
			name:             "solace-api-response",
			server:           httptest.NewServer(mux),
			expectedResponse: []byte(`{"success": true}`),
			expectedErr:      nil,
			expectedBool:     true,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			defer tc.server.Close()
			gotV, gotB, gotErr := CallSolaceSempApi(
				tc.server.URL,
				//"",
				context.TODO(),
				"",
			)
			if gotV != nil || gotB != false || gotErr != nil {
				t.Errorf("got %v, %v,%v", gotV, gotB, gotErr)
			}
		})
	}
}

func TestConstructSempUrl(t *testing.T) {
	n := "0"
	host := "scalable.solace.io/SEMP/v2"
	path := "/api/v1"
	got := ConstructSempUrl(tests.SolaceScalable,
		0,
		"/api/v1",
		map[string]string{
			"param1": "p1==true,p2==false",
			"param2": "p3,p4,p5",
		})
	want := "http://n" + n + "." + host + path + "?" + "param1=p1%3D%3Dtrue,p2%3D%3Dfalse&param2=p3,p4,p5"
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}
}

func TestEncodeForSolace(t *testing.T) {
	s := &SolaceValues{&url.Values{}}
	s.Add("k1", "v1,v2")
	got := s.EncodeForSolace()
	want := "k1=v1,v2"
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}
}
