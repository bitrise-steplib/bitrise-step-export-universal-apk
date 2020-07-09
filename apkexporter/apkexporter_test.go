package apkexporter

import (
	"errors"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-steplib/bitrise-step-generate-universal-apk/bundletool"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_exportAPKs_Successful(t *testing.T) {
	// Given
	mockAPKBuilder := givenMockedAPKBuilder(givenSuccessfulCommand())
	mockFileDownloader := givenMockFileDownloader()
	exporter := givenExporter(mockAPKBuilder, mockFileDownloader)
	aabPath := "/path/to/app.aab"
	tempPath := "/temp/path"
	expectedAPKsPath := "/temp/path/app.apks"

	// When
	output, err := exporter.exportAPKs(aabPath, tempPath, nil)

	// Then
	require.NoError(t, err)
	require.Equal(t, expectedAPKsPath, output)
}

func Test_exportAPKs_FaillingCommand(t *testing.T) {
	// Given
	mockAPKBuilder := givenMockedAPKBuilder(givenFailingCommand())
	mockFileDownloader := givenMockFileDownloader()
	exporter := givenExporter(mockAPKBuilder, mockFileDownloader)

	// When
	output, err := exporter.exportAPKs("", "", nil)

	// Then
	require.Error(t, err)
	require.Empty(t, output)
}

func Test_prepareKeystoreConfig_File(t *testing.T) {
	// Given
	mockAPKBuilder := givenMockedAPKBuilder(givenSuccessfulCommand())
	mockFileDownloader := givenMockFileDownloader()
	exporter := givenExporter(mockAPKBuilder, mockFileDownloader)
	keystoreConfig := givenKeystoreConfig("file://keystore.jks")
	keystoreConfig.KeystorePassword = "pass:password"
	keystoreConfig.SigningKeyPassword = "pass:password"

	// When
	output, err := exporter.prepareKeystoreConfig(keystoreConfig)

	// Then
	require.NoError(t, err)
	require.Contains(t, output.Path, "keystore.jks")
	require.NotContains(t, output.Path, "file:/")
	require.Equal(t, output.KeystorePassword, "pass:password")
	require.Equal(t, output.SigningKeyPassword, "pass:password")
}

func Test_prepareKeystoreConfig_SuccessDownload(t *testing.T) {
	// Given
	mockAPKBuilder := givenMockedAPKBuilder(givenSuccessfulCommand())
	mockFileDownloader := givenMockFileDownloader()
	exporter := givenExporter(mockAPKBuilder, mockFileDownloader)

	// When
	output, err := exporter.prepareKeystoreConfig(givenKeystoreConfig("http://url.com/keystore.jks"))

	// Then
	require.NoError(t, err)
	require.Contains(t, output.Path, "keystore.jks")
	require.Equal(t, output.KeystorePassword, "pass:password")
	require.Equal(t, output.SigningKeyPassword, "pass:password")
}

func Test_prepareKeystoreConfig_FaillingDownload(t *testing.T) {
	// Given
	mockAPKBuilder := givenMockedAPKBuilder(givenSuccessfulCommand())
	expectedError := errors.New("failed")
	mockFileDownloader := givenMockFailingFileDownloader(expectedError)
	exporter := givenExporter(mockAPKBuilder, mockFileDownloader)

	// When
	output, acutalError := exporter.prepareKeystoreConfig(givenKeystoreConfig("http://url.com"))

	// Then
	require.Equal(t, expectedError, acutalError)
	require.Nil(t, output)
}

func Test_apksFilename(t *testing.T) {
	// Given
	aabPath := "/path/to/app.aab"
	expectedAPKSName := "app.apks"

	// When
	actualAPKSName := apksFilename(aabPath)

	// Then
	require.Equal(t, expectedAPKSName, actualAPKSName)
}

func Test_apkFilename(t *testing.T) {
	// Given
	apksPath := "/path/to/app.apks"
	expectedAPKName := "app.apk"

	// When
	actualAPKName := apkFilename(apksPath)

	// Then
	require.Equal(t, expectedAPKName, actualAPKName)
}

func Test_filenameWithExtension(t *testing.T) {
	// Given
	basePath := "/path/to/afile.oldextension"
	expectedFilename := "afile.newextension"

	// When
	actualFilename := filenameWithExtension(basePath, ".newextension")

	// Then
	require.Equal(t, expectedFilename, actualFilename)
}

func Test_keystoreName(t *testing.T) {
	scenarios := []string{
		"https://something.com/debug-keystore.jks",
		"https://something.com/debug-keystore.jks?queryparams",
		"https://something.com/path/debug-keystore.jks",
		"https://something.com/path/debug-keystore.jks?queryparams",
	}

	for _, scenario := range scenarios {
		actualName, err := keystoreName(scenario)

		require.NoError(t, err)
		require.Equal(t, "debug-keystore.jks", actualName)
	}
}

type MockFileDownloader struct {
	mock.Mock
}

func (m *MockFileDownloader) Get(destination, source string) error {
	args := m.Called(destination, source)
	return args.Error(0)
}

func givenMockFileDownloader() *MockFileDownloader {
	mockFileDownloader := new(MockFileDownloader)
	mockFileDownloader.On("Get", mock.Anything, mock.Anything).Return(nil)
	return mockFileDownloader
}

func givenMockFailingFileDownloader(err error) *MockFileDownloader {
	mockFileDownloader := new(MockFileDownloader)
	mockFileDownloader.On("Get", mock.Anything, mock.Anything).Return(err)
	return mockFileDownloader
}

func givenExporter(apkbuilder APKBuilder, filedownloader FileDownloader) Exporter {
	return Exporter{apkbuilder, filedownloader}
}

type MockAPKBuilder struct {
	mock.Mock
}

func (m *MockAPKBuilder) BuildAPKs(aabPath, apksPath string, keystoreCfg *bundletool.KeystoreConfig) *command.Model {
	args := m.Called(aabPath, apksPath, keystoreCfg)
	return args.Get(0).(*command.Model)
}

func givenMockedAPKBuilder(cmd *command.Model) *MockAPKBuilder {
	mockBundletooler := new(MockAPKBuilder)
	mockBundletooler.On("BuildAPKs", mock.Anything, mock.Anything, mock.Anything).Return(cmd)
	return mockBundletooler
}

func givenFailingCommand() *command.Model {
	return command.New("this", "fails")
}

func givenSuccessfulCommand() *command.Model {
	return command.New("echo", "success")
}

func givenKeystoreConfig(path string) *bundletool.KeystoreConfig {
	return &bundletool.KeystoreConfig{
		Path:               path,
		KeystorePassword:   "password",
		SigningKeyAlias:    "alias",
		SigningKeyPassword: "password"}
}
