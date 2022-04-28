//
//  Copyright 2020 The AVFS authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//  	http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//

//go:build mage

// avfs is the build script for AVFS.
package main

import (
	"fmt"
	"go/build"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	dockerGoSrc = "/go/src"
	dockerImage = "avfs-docker"
	gitCmd      = "git"
	goCmd       = "go"
	goFumptCmd  = "gofumpt"
	goFumptUrl  = "mvdan.cc/gofumpt@master"
	golangCiCmd = "golangci-lint"
	golangCiGit = "github.com/golangci/golangci-lint"
	golangCiBin = "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"
	goxCmd      = "gox"
	goxUrl      = "github.com/mitchellh/gox@master"
	raceCount   = 5
	benchCount  = 5
)

var (
	appDir            string
	dockerCmd         string
	tmpDir            string
	coverTestPath     string
	coverRacePath     string
	testDataDir       string
	dockerTmpDir      string
	dockerTestDataDir string
)

func init() {
	appDir, _ = os.Getwd()
	appDir = strings.TrimSuffix(appDir, "mage")

	tmpDir = filepath.Join(appDir, "tmp")
	coverTestPath = filepath.Join(tmpDir, "cover_test.txt")
	coverRacePath = filepath.Join(tmpDir, "cover_race.txt")
	testDataDir = filepath.Join(appDir, "test/testdata")

	var dockerVolume string
	if runtime.GOOS == "windows" {
		dockerVolume = "c:"
	}

	dockerTmpDir = filepath.Join(dockerVolume, dockerGoSrc, "tmp")
	dockerTestDataDir = filepath.Join(dockerVolume, dockerGoSrc, "test/testdata")

	switch {
	case isExecutable("docker"):
		dockerCmd = "docker"
	case isExecutable("podman"):
		dockerCmd = "podman"
	default:
		dockerCmd = ""
	}
}

// tmpInit creates the temporary directory.
func tmpInit() error {
	_, err := os.Stat(tmpDir)
	if err == nil {
		return nil
	}

	err = os.MkdirAll(tmpDir, 0o755)
	if err != nil {
		return err
	}

	return os.Chmod(tmpDir, 0o777)
}

// Env returns the go environment variables.
func Env() {
	sh.RunV(goCmd, "env")
	fmt.Printf(`
appDir=%s
tmpDir=%s
coverTestPath=%s
coverRacePath=%s
testDataDir=%s
dockerTmpDir=%s
dockerTestDataDir=%s
`,
		appDir, tmpDir, coverTestPath, coverRacePath,
		testDataDir, dockerTmpDir, dockerTestDataDir)
}

// Build builds the project.
func Build() error {
	return sh.RunV(goCmd, "build", "-v", "./...")
}

// Fmt runs gofumpt on the project.
func Fmt() error {
	if !isExecutable(goFumptCmd) {
		err := os.Chdir(os.TempDir())
		if err != nil {
			return err
		}

		err = sh.RunV(goCmd, "install", goFumptUrl)
		if err != nil {
			return err
		}

		err = os.Chdir(appDir)
		if err != nil {
			return err
		}
	}

	return sh.RunV(goFumptCmd, "-l", "-w", "-extra", ".")
}

// Lint runs golangci-lint (on Windows it must be run from a bash shell like git bash).
func Lint() error {
	if !isExecutable(golangCiCmd) {
		version, err := gitLastVersion(golangCiGit)
		if err != nil {
			return err
		}

		fmt.Printf("version = %s\n", version)

		script := filepath.Join(os.TempDir(), golangCiCmd+".sh")

		err = downloadFile(script, golangCiBin)
		if err != nil {
			return err
		}

		defer os.Remove(script)

		binDir := filepath.Join(build.Default.GOPATH, "bin")

		err = sh.RunV("sh", script, "-b", binDir, version)
		if err != nil {
			return err
		}
	}

	return sh.RunV(golangCiCmd, "run", "-v")
}

// CoverResult opens a web browser with the latest coverage file.
func CoverResult() error {
	if isCI() {
		return nil
	}

	return sh.RunV(goCmd, "tool", "cover", "-html="+coverTestPath)
}

// Test runs tests with coverage.
func Test() error {
	mg.Deps(tmpInit)

	err := sh.RunV(goCmd, "test",
		"-run=.",
		"-race", "-v",
		"-covermode=atomic",
		"-coverprofile="+coverTestPath,
		"./...")
	if err != nil {
		return err
	}

	return CoverResult()
}

// TestBuild builds a test executable on all architectures (except Android/*)
func TestBuild() error {
	mg.Deps(tmpInit)

	if !isExecutable(goxCmd) {
		err := sh.RunV(goCmd, "install", goxUrl)
		if err != nil {
			return err
		}
	}

	osArch, err := sh.Output(goCmd, "tool", "dist", "list")
	if err != nil {
		return err
	}

	// Remove Android platforms : need additional tools.
	re := regexp.MustCompile("android/[^\n]+\n")
	osArch = re.ReplaceAllString(osArch, "")
	osArch = strings.ReplaceAll(osArch, "\n", " ")

	srcPath := filepath.Join(appDir, "test/testbuild")
	outPath := filepath.Join(appDir, "tmp/{{.Dir}}/{{.Dir}}_{{.OS}}_{{.Arch}}")

	err = sh.RunV(goxCmd, "-cgo",
		"-osarch=\""+osArch+"\"",
		"-output="+outPath,
		srcPath)
	if err != nil {
		return err
	}

	return nil
}

// Race runs data race tests.
func Race() error {
	mg.Deps(tmpInit)

	return sh.RunV(goCmd, "test",
		"-tags=datarace",
		"-run=TestRace",
		"-race", "-v",
		"-count="+strconv.Itoa(raceCount),
		"-covermode=atomic",
		"-coverprofile="+coverRacePath,
		"./...")
}

// Bench runs benchmarks.
func Bench() error {
	return sh.RunV(goCmd, "test",
		"-run=^a",
		"-bench=.",
		"-benchmem",
		"-count="+strconv.Itoa(benchCount),
		"./...")
}

// DockerBuild builds docker image for AVFS.
func DockerBuild() error {
	mg.Deps(tmpInit)

	if dockerCmd == "" {
		return fmt.Errorf("can't find docker or podman in the current path")
	}

	var (
		image string
		user  string
	)

	err := sh.RunV("tar",
		"-cf", "tmp/avfs.tar",
		"--exclude-vcs",
		"--exclude-ignore='.gitignore'",
		".")
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		image = "golang:windowsservercore"
		user = "ContainerAdministrator"
	case "linux":
		image = "golang:bullseye"
		user = "root"
	}

	fmt.Printf("image = %s\nuser = %s\n", image, user)

	return sh.RunV(dockerCmd,
		"build",
		"-t", dockerImage,
		"--build-arg", "image="+image,
		"--build-arg", "user="+user,
		".")
}

// DockerTerm opens a shell as root in the docker image for AVFS.
func DockerTerm() error {
	mg.Deps(DockerBuild)

	shell := "bash"
	if runtime.GOOS == "windows" {
		shell = "cmd"
	}

	return dockerTest(shell)
}

// DockerTest runs tests in the docker image and displays the coverage result.
func DockerTest() error {
	mg.Deps(DockerBuild)

	err := dockerTest()
	if err != nil {
		return err
	}

	return CoverResult()
}

// DockerPrune removes unused data from Docker.
func DockerPrune() error {
	return sh.RunV(dockerCmd, "system", "prune", "-f")
}

// dockerTest runs tests in the docker image for AVFS.
func dockerTest(args ...string) error {
	termOptions := "-it"
	if runtime.GOOS == "windows" {
		termOptions = "-i"
	}

	tmpMount := tmpDir + ":" + dockerTmpDir
	testDataMount := testDataDir + ":" + dockerTestDataDir
	cmdArgs := []string{
		"run",
		termOptions,
		"-v", tmpMount,
		"-v", testDataMount,
		dockerImage,
	}

	cmdArgs = append(cmdArgs, args...)

	return sh.RunV(dockerCmd, cmdArgs...)
}

// isExecutable checks if name is an executable in the current path.
func isExecutable(name string) bool {
	_, err := exec.LookPath(name)

	return err == nil
}

// isCI tests if we run in a CI environment.
func isCI() bool {
	return os.Getenv("CI") != ""
}

// gitLastVersion return the latest tagged version of a remote git repository.
func gitLastVersion(repo string) (string, error) {
	const semverRegexp = "v\\d+\\.\\d+\\.\\d+$"

	if !strings.HasPrefix(repo, "https://") {
		repo = "https://" + repo
	}

	out, err := sh.Output(gitCmd, "ls-remote",
		"--tags",
		"--refs",
		"--sort=v:refname",
		repo)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(semverRegexp)

	version := re.FindString(out)
	if version == "" {
		return "", fmt.Errorf("version : incorrect format :\n%s", out)
	}

	return version, nil
}

// downloadFile downloads a url to a local file.
func downloadFile(path, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.Copy(f, resp.Body)

	return err
}
