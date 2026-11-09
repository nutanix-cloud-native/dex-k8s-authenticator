package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mesosphere/dex-k8s-authenticator/pkg/tenancy"
)

func pluginTestConfig(multiTenancy bool) *Config {
	return &Config{
		Enable_Multi_Tenancy: multiTenancy,
		Web_Path_Prefix:      "/token/",
		Clusters: []Cluster{
			{
				Name:           tenancy.ManagementClusterName,
				Issuer:         "https://dex.example.com",
				Redirect_URI:   "https://example.com/token/callback",
				K8s_Master_URI: "https://mgmt.example.com:6443",
				TenantId:       tenancy.TenantId(tenancy.ManagementWorkspaceName),
			},
			{
				Name:           "dev-cluster-01",
				Issuer:         "https://dex.example.com",
				Redirect_URI:   "https://example.com/token/callback",
				K8s_Master_URI: "https://dev.example.com:6443",
				TenantId:       "york-ws",
			},
		},
	}
}

func TestDownloadUrl(t *testing.T) {
	url := "https://www.example.com/token/"

	testCases := []struct {
		platform    string
		version     string
		binary      string
		expectedURL string
	}{
		{
			platform:    "linux_amd64",
			version:     "v0.1.0",
			binary:      "konvoy-async-auth",
			expectedURL: "https://www.example.com/token/static/downloads/konvoy-async-auth_v0.1.0_linux_amd64/konvoy-async-auth",
		},
		{
			platform:    "darwin_arm64",
			version:     "v0.2.0",
			binary:      "konvoy-async-auth",
			expectedURL: "https://www.example.com/token/static/downloads/konvoy-async-auth_v0.2.0_darwin_arm64/konvoy-async-auth",
		},
		{
			platform:    "windows_amd64",
			version:     "v0.2.1",
			binary:      "konvoy-async-auth.exe",
			expectedURL: "https://www.example.com/token/static/downloads/konvoy-async-auth_v0.2.1_windows_amd64/konvoy-async-auth.exe",
		},
	}

	for _, tc := range testCases {
		downloadURL := getDownloadURL(url, tc.platform, tc.version, tc.binary)
		assert.Equal(t, downloadURL, tc.expectedURL)
	}
}

func TestGetClustersByProviders_MultiTenancyMissingTenantID(t *testing.T) {
	config := pluginTestConfig(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/plugin/providers", nil)

	config.getClustersByProviders(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing tenant-id")
	assert.NotContains(t, rr.Body.String(), "dev-cluster-01")
}

func TestGetClustersByProviders_FiltersByTenantID(t *testing.T) {
	config := pluginTestConfig(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/plugin/providers?tenant-id=york-ws", nil)

	config.getClustersByProviders(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, []string{"dev-cluster-01"}, providerClusterNames(t, rr.Body.Bytes()))
}

func TestGetClustersByProviders_ManagementWorkspaceIncludesKubernetesCluster(t *testing.T) {
	config := pluginTestConfig(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/plugin/providers?tenant-id="+tenancy.ManagementWorkspaceName, nil)

	config.getClustersByProviders(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, []string{tenancy.ManagementClusterName}, providerClusterNames(t, rr.Body.Bytes()))
}

func TestGetClustersByProviders_EmptyWorkspaceReturnsEmptyArray(t *testing.T) {
	config := pluginTestConfig(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/plugin/providers?tenant-id=empty-ws", nil)

	config.getClustersByProviders(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "[]", rr.Body.String())
}

func TestGetClustersByProviders_MultiTenancyOffReturnsAll(t *testing.T) {
	config := pluginTestConfig(false)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/plugin/providers", nil)

	config.getClustersByProviders(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.ElementsMatch(t, []string{tenancy.ManagementClusterName, "dev-cluster-01"}, providerClusterNames(t, rr.Body.Bytes()))
}

func TestDownloadKubeConfig_MultiTenancyMissingTenantID(t *testing.T) {
	config := pluginTestConfig(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/plugin/kubeconfig", nil)

	config.downloadKubeConfigUnix(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing tenant-id")
}

func TestDownloadKubeConfigWindows_MultiTenancyMissingTenantID(t *testing.T) {
	config := pluginTestConfig(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/plugin/kubeconfig_windows", nil)

	config.downloadKubeConfigWindows(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing tenant-id")
}

func TestDownloadKubeConfig_FiltersByTenantID(t *testing.T) {
	config := pluginTestConfig(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/plugin/kubeconfig?tenant-id=york-ws", nil)

	config.downloadKubeConfigUnix(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(t, body, "dev.example.com")
	assert.NotContains(t, body, "mgmt.example.com")
	assert.Contains(t, body, "command: "+binaryName)
}

func TestRenderInstructions_MultiTenancyMissingTenantID(t *testing.T) {
	config := pluginTestConfig(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/plugin/data", nil)

	config.renderInstructions(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing tenant-id")
}

func TestRenderInstructions_ClusterQueryWithoutTenantID(t *testing.T) {
	config := pluginTestConfig(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/plugin/data?cluster=dev.example.com", nil)

	config.renderInstructions(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing tenant-id")
}

func TestRenderInstructions_ClusterQueryWrongTenant(t *testing.T) {
	config := pluginTestConfig(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/token/plugin/data?tenant-id=york-ws&cluster=mgmt.example.com", nil)

	config.renderInstructions(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestClusterFilterFromRequest(t *testing.T) {
	t.Parallel()

	mgmt := Cluster{Name: tenancy.ManagementClusterName, TenantId: tenancy.TenantId(tenancy.ManagementWorkspaceName)}
	attached := Cluster{Name: "dev-cluster-01", TenantId: "york-ws"}

	t.Run("multi-tenancy missing tenant-id", func(t *testing.T) {
		t.Parallel()
		config := pluginTestConfig(true)
		req := httptest.NewRequest(http.MethodGet, "/token/plugin/providers", nil)
		_, err := config.clusterFilterFromRequest(req)
		assert.ErrorIs(t, err, errMissingTenantID)
	})

	t.Run("multi-tenancy filters by tenant-id", func(t *testing.T) {
		t.Parallel()
		config := pluginTestConfig(true)
		req := httptest.NewRequest(http.MethodGet, "/token/plugin/providers?tenant-id=york-ws", nil)
		filter, err := config.clusterFilterFromRequest(req)
		require.NoError(t, err)
		assert.False(t, filter(mgmt))
		assert.True(t, filter(attached))
	})

	t.Run("multi-tenancy off includes all", func(t *testing.T) {
		t.Parallel()
		config := pluginTestConfig(false)
		req := httptest.NewRequest(http.MethodGet, "/token/plugin/providers", nil)
		filter, err := config.clusterFilterFromRequest(req)
		require.NoError(t, err)
		assert.True(t, filter(mgmt))
		assert.True(t, filter(attached))
	})
}

func providerClusterNames(t *testing.T, body []byte) []string {
	t.Helper()
	var providers []FlatProviderMap
	require.NoError(t, json.Unmarshal(body, &providers))
	var names []string
	for _, p := range providers {
		for _, c := range p.Clusters {
			names = append(names, c.Name)
		}
	}
	return names
}
