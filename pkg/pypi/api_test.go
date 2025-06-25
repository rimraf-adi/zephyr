package pypi

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchPackageMetadata_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"info": {"name": "foo", "version": "1.0.0"}, "releases": {}, "urls": []}`))
	}))
	defer ts.Close()
	client := &PyPIClient{httpClient: ts.Client(), baseURL: ts.URL}
	meta, err := client.FetchPackageMetadata("foo")
	if err != nil {
		t.Fatalf("FetchPackageMetadata failed: %v", err)
	}
	if meta.Info.Name != "foo" || meta.Info.Version != "1.0.0" {
		t.Errorf("Metadata mismatch: %+v", meta.Info)
	}
}

func TestFetchPackageMetadata_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()
	client := &PyPIClient{httpClient: ts.Client(), baseURL: ts.URL}
	_, err := client.FetchPackageMetadata("foo")
	if err == nil {
		t.Error("Expected error for HTTP 404, got nil")
	}
}

func TestFetchSimpleIndex_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>simple index</body></html>"))
	}))
	defer ts.Close()
	client := &PyPIClient{httpClient: ts.Client(), baseURL: ts.URL}
	body, err := client.FetchSimpleIndex("foo")
	if err != nil || !strings.Contains(body, "simple index") {
		t.Errorf("FetchSimpleIndex failed: %v, body=%s", err, body)
	}
}

func TestGetLatestVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"info": {"name": "foo", "version": "2.0.0"}, "releases": {}, "urls": []}`))
	}))
	defer ts.Close()
	client := &PyPIClient{httpClient: ts.Client(), baseURL: ts.URL}
	ver, err := client.GetLatestVersion("foo")
	if err != nil || ver != "2.0.0" {
		t.Errorf("GetLatestVersion failed: %v, ver=%s", err, ver)
	}
}

func TestGetVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"info": {"name": "foo", "version": "1.0.0"}, "releases": {"1.0.0": [], "2.0.0": []}, "urls": []}`))
	}))
	defer ts.Close()
	client := &PyPIClient{httpClient: ts.Client(), baseURL: ts.URL}
	vers, err := client.GetVersions("foo")
	if err != nil || len(vers) != 2 {
		t.Errorf("GetVersions failed: %v, vers=%v", err, vers)
	}
}

func TestGetReleasesForVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"info": {"name": "foo", "version": "1.0.0"}, "releases": {"1.0.0": [{"filename": "foo-1.0.0.whl", "url": "http://example.com", "size": 123, "upload_time": "2024-01-01T00:00:00", "digests": {"sha256": "abc"}, "python_version": "py3", "packagetype": "bdist_wheel"}]}, "urls": []}`))
	}))
	defer ts.Close()
	client := &PyPIClient{httpClient: ts.Client(), baseURL: ts.URL}
	rels, err := client.GetReleasesForVersion("foo", "1.0.0")
	if err != nil || len(rels) != 1 {
		t.Errorf("GetReleasesForVersion failed: %v, rels=%v", err, rels)
	}
}

func TestDownloadRelease(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("wheel content"))
	}))
	defer ts.Close()
	rel := Release{URL: ts.URL}
	client := &PyPIClient{httpClient: ts.Client(), baseURL: ts.URL}
	rc, err := client.DownloadRelease(rel)
	if err != nil {
		t.Fatalf("DownloadRelease failed: %v", err)
	}
	data, _ := ioutil.ReadAll(rc)
	if string(data) != "wheel content" {
		t.Errorf("DownloadRelease content mismatch: %s", string(data))
	}
	rc.Close()
}

func TestFindWheelForVersion(t *testing.T) {
	client := &PyPIClient{}
	releases := []Release{
		{Filename: "foo-1.0.0.whl", Packagetype: "bdist_wheel"},
		{Filename: "foo-1.0.0.tar.gz", Packagetype: "sdist"},
	}
	// Simulate GetReleasesForVersion
	client.GetReleasesForVersion = func(pkg, ver string) ([]Release, error) {
		return releases, nil
	}
	rel, err := client.FindWheelForVersion("foo", "1.0.0", "any")
	if err != nil || rel.Filename != "foo-1.0.0.whl" {
		t.Errorf("FindWheelForVersion failed: %v, rel=%+v", err, rel)
	}
} 