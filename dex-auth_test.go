package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mesosphere/dex-k8s-authenticator/pkg/tenancy"
)

func TestHandleIndex_MultiTenancyRedirectsToManagementLogin(t *testing.T) {
	config := pluginTestConfig(true)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/", nil)
	config.handleIndex(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Equal(t, "/token/login/"+tenancy.ManagementClusterName, rr.Header().Get("Location"))
	assert.NotContains(t, rr.Body.String(), "dev-cluster-01")
	assert.NotContains(t, rr.Body.String(), "Select which cluster")
}

func TestHandleIndex_MultiTenancyIgnoresClusterCount(t *testing.T) {
	config := pluginTestConfig(true)
	config.Clusters = config.Clusters[:1]

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/", nil)
	config.handleIndex(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Equal(t, "/token/login/"+tenancy.ManagementClusterName, rr.Header().Get("Location"))
}

func TestHandleIndex_SingleClusterRedirectsWhenMultiTenancyOff(t *testing.T) {
	config := pluginTestConfig(false)
	config.Clusters = []Cluster{
		{Name: "example-cluster"},
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/", nil)
	config.handleIndex(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Equal(t, "/token/login/example-cluster", rr.Header().Get("Location"))
}

func TestHandleIndex_MultipleClustersRendersPickerWhenMultiTenancyOff(t *testing.T) {
	config := pluginTestConfig(false)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/", nil)
	config.handleIndex(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), tenancy.ManagementClusterName)
	assert.Contains(t, rr.Body.String(), "dev-cluster-01")
	assert.Contains(t, rr.Body.String(), "Select which cluster")
}
