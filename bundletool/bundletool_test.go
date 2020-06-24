package bundletool

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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
	expectedCommand := buildAPKsCommand(tool, aabPath, apksPath, nil)

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
	expectedCommand := buildAPKsCommand(tool, aabPath, apksPath, &keystoreConfig)

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
	actualSources, err := sources(version)

	//Then
	require.NoError(t, err)
	require.Equal(t, expectedSources, actualSources)
}

func Test_New_Success(t *testing.T) {
	// Given
	mockedFileDownloader := new(MockFileDownloader)
	mockedFileDownloader.On("GetWithFallback", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// When
	tool, err := New("0.2.0", mockedFileDownloader)

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
	tool, actualError := New("0.2.0", mockedFileDownloader)

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

func buildAPKsCommand(tool Tool, aabPath, apksPath string, keystoreCfg *KeystoreConfig) []string {
	command := append([]string{"java", "-jar", tool.path, "build-apks"})
	command = append(command, "--mode=universal")
	command = append(command, "--bundle", aabPath)
	command = append(command, "--output", apksPath)

	if keystoreCfg != nil {
		command = append(command, "--ks", keystoreCfg.Path)
		command = append(command, "--ks-pass", keystoreCfg.KeystorePassword)
		command = append(command, "--ks-key-alias", keystoreCfg.SigningKeyAlias)
		command = append(command, "--key-pass", keystoreCfg.SigningKeyPassword)
	}

	return command
}

func assertFileExists(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("File should exist at: %s", path)
	}
}
