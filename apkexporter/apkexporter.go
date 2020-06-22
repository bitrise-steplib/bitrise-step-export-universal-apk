package apkexporter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-steplib/bitrise-step-generate-universal-apk/bundletool"
	"github.com/bitrise-steplib/steps-deploy-to-bitrise-io/androidartifact"
)

// Bundletooler ...
type Bundletooler interface {
	BuildAPKs(aabPath, apksPath string, keystoreCfg *bundletool.KeystoreConfig) *command.Model
}

// Exporter ...
type Exporter struct {
	bundletooler Bundletooler
}

// New ...
func New(bundletooler Bundletooler) Exporter {
	return Exporter{bundletooler: bundletooler}
}

// unzipUniversalAPKsArchive unzips an universal apks archive.
func unzipUniversalAPKsArchive(archive, destDir string) (string, error) {
	unzipCommand := command.New("unzip", archive, "-d", destDir)
	return filepath.Join(destDir, "universal.apk"), run(unzipCommand)
}

// handleError creates error with layout: `<cmd> failed (status: <status_code>): <cmd output>`.
func handleError(cmd, out string, err error) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf("%s failed", cmd)
	if status, exitCodeErr := errorutil.CmdExitCodeFromError(err); exitCodeErr == nil {
		msg += fmt.Sprintf(" (status: %d)", status)
	}
	if len(out) > 0 {
		msg += fmt.Sprintf(": %s", out)
	}
	return errors.New(msg)
}

// run executes a given command.
func run(cmd *command.Model) error {
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	return handleError(cmd.PrintableCommandArgs(), out, err)
}

// ExportUniversalAPK generates universal apks from an aab file.
func (exporter Exporter) ExportUniversalAPK(aabPath, destDir string, keystoreConfig *bundletool.KeystoreConfig) (string, error) {
	tempPath, err := pathutil.NormalizedOSTempDirPath("universal_apk")
	if err != nil {
		return "", err
	}

	apksPath, err := exporter.exportAPKs(aabPath, tempPath, keystoreConfig)
	if err != nil {
		return "", err
	}

	universalAPKPath, err := unzipUniversalAPKsArchive(apksPath, tempPath)
	if err != nil {
		return "", err
	}

	renamedUniversalAPKPath := filepath.Join(tempPath, universalAPKName(aabPath))
	err = os.Rename(universalAPKPath, renamedUniversalAPKPath)
	if err != nil {
		return "", err
	}

	err = command.CopyFile(renamedUniversalAPKPath, filepath.Join(destDir, filepath.Base(renamedUniversalAPKPath)))
	if err != nil {
		return "", err
	}

	return universalAPKPath, nil
}

func universalAPKName(aabPath string) string {
	untrimmedAPKName := androidartifact.UniversalAPKBase(aabPath)
	extension := filepath.Ext(untrimmedAPKName)
	fileNameWithoutExtension := strings.TrimSuffix(untrimmedAPKName, extension)
	trimmedFileName := strings.Trim(fileNameWithoutExtension, "-")
	return trimmedFileName + extension
}

func (exporter Exporter) exportAPKs(aabPath, tempPath string, keystoreConfig *bundletool.KeystoreConfig) (string, error) {
	apksPath := filepath.Join(tempPath, apksFilename(aabPath))

	buildAPKsCommand := exporter.bundletooler.BuildAPKs(aabPath, apksPath, keystoreConfig)
	err := run(buildAPKsCommand)
	if err != nil {
		return "", err
	}

	return apksPath, nil
}

func apksFilename(aabPath string) string {
	return filenameWithExtension(aabPath, ".apks")
}

func apkFilename(apksPath string) string {
	return filenameWithExtension(apksPath, ".apk")
}

func filenameWithExtension(basepath, extension string) string {
	filename := filepath.Base(basepath)
	fileNameWithoutExtension := strings.TrimSuffix(filename, filepath.Ext(filename))
	return fileNameWithoutExtension + extension
}
