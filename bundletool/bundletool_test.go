package bundletool

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	logv2 "github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/retryhttp"
	"github.com/bitrise-steplib/bitrise-step-export-universal-apk/filedownloader"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Bundletool_Download(t *testing.T) {
	bundletoolversion := "1.15.4"
	downloader := filedownloader.New(retryhttp.NewClient(logv2.NewLogger()))

	reqNum := 0
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/1.15.4/bundletool-all-1.15.4.jar" {
			if reqNum == 0 {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
			reqNum++
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

	}))
	defer svr.Close()

	_, err := New(bundletoolversion, downloader, svr.URL)
	require.NoError(t, err)
}

func Test_BuildCommand(t *testing.T) {
	// Given
	tool := givenTool()
	cmd := "command"
	args := []string{"arg1", "arg2"}
	expectedCommand := append([]string{"java", "-jar", tool.path, cmd}, args...)

	// When
	actualCommand := tool.BuildCommand(cmd, args...).GetCmd().Args

	// Then
	require.Equal(t, expectedCommand, actualCommand)
}

func Test_BuildAPKs_withoutKeystoreConfig(t *testing.T) {
	// Given
	tool := givenTool()
	aabPath := "/path/to/app.aab"
	apksPath := "/path/to/app.apks"
	expectedCommand := []string{"java", "-jar", tool.path, "build-apks", "--mode=universal", "--bundle", aabPath, "--output", apksPath}

	// When
	actualCommand := tool.BuildAPKs(aabPath, apksPath, nil).GetCmd().Args

	// Then
	require.Equal(t, expectedCommand, actualCommand)
}

func Test_BuildAPKs_withKeystoreConfig(t *testing.T) {
	// Given
	tool := givenTool()
	aabPath := "/path/to/app.aab"
	apksPath := "/path/to/app.apks"
	keystoreConfig := givenKeystoreConfig()
	expectedCommand := []string{"java", "-jar", tool.path, "build-apks", "--mode=universal", "--bundle", aabPath, "--output", apksPath,
		"--ks", keystoreConfig.Path, "--ks-pass", keystoreConfig.KeystorePassword, "--ks-key-alias", keystoreConfig.SigningKeyAlias,
		"--key-pass", keystoreConfig.SigningKeyPassword}

	// When
	actualCommand := tool.BuildAPKs(aabPath, apksPath, &keystoreConfig).GetCmd().Args

	// Then
	require.Equal(t, expectedCommand, actualCommand)
}

func Test_sources(t *testing.T) {
	// Given
	version := "0.1.0"
	expectedSources := []string{
		"https://github.com/google/bundletool/releases/download/0.1.0/bundletool-all-0.1.0.jar",
		"https://github.com/google/bundletool/releases/download/0.1.0/bundletool-all.jar",
	}

	// When
	actualSources, err := sources(version, GithubReleaseBaseURL)

	//Then
	require.NoError(t, err)
	require.Equal(t, expectedSources, actualSources)
}

func Test_New_Success(t *testing.T) {
	// Given
	mockedFileDownloader := new(MockFileDownloader)
	mockedFileDownloader.On("GetWithFallback", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// When
	tool, err := New("0.2.0", mockedFileDownloader, GithubReleaseBaseURL)

	// Then
	require.NoError(t, err)
	require.NotNil(t, tool)
}

func Test_New_Fail(t *testing.T) {
	// Given
	mockedFileDownloader := new(MockFileDownloader)
	expectedError := errors.New("failed")
	mockedFileDownloader.On("GetWithFallback", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	// When
	tool, actualError := New("0.2.0", mockedFileDownloader, GithubReleaseBaseURL)

	// Then
	require.Equal(t, expectedError, actualError)
	require.Nil(t, tool)
}

type MockFileDownloader struct {
	mock.Mock
}

func (m *MockFileDownloader) GetWithFallback(destination, source string, fallbackSources ...string) error {
	args := m.Called(destination, source, fallbackSources)
	return args.Error(0)
}

func givenTool() Tool {
	return Tool{"/whatever/path"}
}

func givenKeystoreConfig() KeystoreConfig {
	return KeystoreConfig{Path: "/path/to/keystore.keystore",
		KeystorePassword:   "pass:keystorePassword",
		SigningKeyAlias:    "signingkeyalias",
		SigningKeyPassword: "file:/path/to/keystorepassfile"}
}

func assertFileExists(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("File should exist at: %s", path)
	}
}
