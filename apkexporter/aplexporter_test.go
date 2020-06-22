package apkexporter

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-steplib/bitrise-step-generate-universal-apk/bundletool"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_exportAPKs_Successful(t *testing.T) {
	// Given
	mockBundletooler := givenMockedBundletooler(givenSuccessfulCommand())
	exporter := givenExporter(mockBundletooler)
	aabPath := "/path/to/app.aab"
	tempPath := "/temp/path"
	expectedAPKsPath := "/temp/path/app.apks"

	// When
	output, err := exporter.exportAPKs(aabPath, tempPath, nil)

	// Then
	require.Equal(t, expectedAPKsPath, output)
	require.Nil(t, err)
}

func Test_exportAPKs_Failling(t *testing.T) {
	// Given
	mockBundletooler := givenMockedBundletooler(givenFailingCommand())
	exporter := givenExporter(mockBundletooler)

	// When
	output, err := exporter.exportAPKs("", "", nil)

	// Then
	require.Empty(t, output)
	require.NotNil(t, err)
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

func givenExporter(bundletooler Bundletooler) Exporter {
	return Exporter{bundletooler}
}

type MockBundletooler struct {
	mock.Mock
}

func (m *MockBundletooler) BuildAPKs(aabPath, apksPath string, keystoreCfg *bundletool.KeystoreConfig) *command.Model {
	args := m.Called(aabPath, apksPath, keystoreCfg)
	return args.Get(0).(*command.Model)
}

func givenMockedBundletooler(cmd *command.Model) *MockBundletooler {
	mockBundletooler := new(MockBundletooler)
	mockBundletooler.On("BuildAPKs", mock.Anything, mock.Anything, mock.Anything).Return(cmd)
	return mockBundletooler
}

func givenFailingCommand() *command.Model {
	return command.New("this", "fails")
}

func givenSuccessfulCommand() *command.Model {
	return command.New("echo", "success")
}
