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
	mockAPKBuilder := givenMockedAPKBuilder(givenSuccessfulCommand())
	exporter := givenExporter(mockAPKBuilder)
	aabPath := "/path/to/app.aab"
	tempPath := "/temp/path"
	expectedAPKsPath := "/temp/path/app.apks"

	// When
	output, err := exporter.exportAPKs(aabPath, tempPath, nil)

	// Then
	require.NoError(t, err)
	require.Equal(t, expectedAPKsPath, output)
}

func Test_exportAPKs_Failling(t *testing.T) {
	// Given
	mockAPKBuilder := givenMockedAPKBuilder(givenFailingCommand())
	exporter := givenExporter(mockAPKBuilder)

	// When
	output, err := exporter.exportAPKs("", "", nil)

	// Then
	require.Error(t, err)
	require.Empty(t, output)
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

func givenExporter(apkbuilder APKBuilder) Exporter {
	return Exporter{apkbuilder}
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
