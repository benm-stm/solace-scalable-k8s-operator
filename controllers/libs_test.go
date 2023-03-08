package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLabels(t *testing.T) {
	got := Labels(&solaceScalable)
	want := map[string]string{
		"app": "test",
	}
	if got["app"] != want["app"] {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestIsItInSlice(t *testing.T) {
	got := IsItInSlice("a", []string{"a", "b", "c"})
	want := true
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}
}

func TestAsSha256(t *testing.T) {
	got := AsSha256("test")
	want := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}

func TestUniqueAndNonZero(t *testing.T) {
	got := UniqueAndNonZero([]int32{1, 2, 3, 2, 0})
	want := []int32{1, 2, 3}
	if len(got) != len(want) {
		t.Errorf("got %v, wanted %v", got, want)
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("got %v, wanted %v", got, want)
		}
	}
}

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
			if gotV != "" || gotB != false || gotErr != nil {
				t.Errorf("got %v, %v,%v", gotV, gotB, gotErr)
			}
		})
	}
}

func TestConstructSempUrl(t *testing.T) {
	got := ConstructSempUrl(solaceScalable,
		0,
		"/api/v1",
		map[string]string{
			"param1": "p1==true,p2==false",
			"param2": "p3,p4,p5",
		})
	want := "http://n0.scalable.solace.io/SEMP/v2/api/v1?param1=p1%3D%3Dtrue,p2%3D%3Dfalse&param2=p3,p4,p5"
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}
}

func TestContains(t *testing.T) {
	got := Contains([]string{"a", "b", "c", "a"},
		"a",
	)
	want := true
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}
}

func TestNextAvailablePort(t *testing.T) {
	got := NextAvailablePort([]int32{1025, 1026, 1028, 1030},
		1025,
	)
	want := 1027
	if int(got) != want {
		t.Errorf("got %v, wanted %v", got, want)
	}
}

func TestReformatForSolace(t *testing.T) {
	got := ReformatForSolace("abc%2Cab%2Ccab")
	want := "abc,ab,cab"
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}
}
