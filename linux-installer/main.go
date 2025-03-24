package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	controlTemplatePath  = "./linux-installer/deb-templates/control.template"
	postinstTemplatePath = "./linux-installer/deb-templates/postinst.template"
	prermTemplatePath    = "./linux-installer/deb-templates/prerm.template"
	specTemplatePath     = "./linux-installer/rpm-templates/package.spec.template"
)

const (
	buildDir   = "./linux-installer/build"
	packageDir = "./linux-installer/package"

	commonInstallDir     = "/opt/packet-sentry"
	commonSystemdDir     = "/etc/systemd/system"
	commonBinDir         = commonInstallDir + "/bin"
	commonBinFile        = commonBinDir + "/packet-sentry-agent"
	commonSystemdSvcFile = "/packet-sentry-agent.service"

	debBuildDir     = buildDir + "/deb"
	debDebianDir    = debBuildDir + "/DEBIAN"
	debControlPath  = debBuildDir + "/DEBIAN/control"
	debPostinstPath = debBuildDir + "/DEBIAN/postinst"
	debPrermPath    = debBuildDir + "/DEBIAN/prerm"
	debInstallDir   = debBuildDir + commonInstallDir
	debSystemdDir   = debBuildDir + commonSystemdDir
	debBinDir       = debInstallDir + "/bin"
)

type PackageInfo struct {
	Architecture    string
	PackageName     string
	PackageVersion  string
	BootstrapFile   string
	InstallDir      string
	MaintainerEmail string
	BinFile         string
}

func readFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filePath, err)
	}
	return string(content), nil
}

func writeToFile(filePath, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}
	return nil
}

func processTemplate(templateContent string, data interface{}) (string, error) {
	tmpl, err := template.New("template").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	var builder strings.Builder
	err = tmpl.Execute(&builder, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}
	return builder.String(), nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Invalid number of arguments.\n\n\tUsage: go run main.go [version] [arch]\n\n")
		os.Exit(1)
	}

	version := os.Args[1]
	arch := os.Args[2]

	goBuildBinary := fmt.Sprintf("./build/packet_sentry_linux_%s", arch)
	if _, err := os.Stat(goBuildBinary); os.IsNotExist(err) {
		fmt.Printf("Error: Binary %s not found. Run `./scripts/build linux %s` to build it.\n", goBuildBinary, arch)
		os.Exit(1)
	}

	// DEB package setup
	debPackageInfo := PackageInfo{
		PackageName:     "packet-sentry-agent",
		PackageVersion:  version,
		BootstrapFile:   commonInstallDir + "/agentBootstrap.json",
		InstallDir:      commonInstallDir,
		MaintainerEmail: "maintainer@example.com",
		Architecture:    arch,
		BinFile:         commonBinFile,
	}

	requiredDebDirs := []string{
		debBuildDir,
		debDebianDir,
		debInstallDir,
		debBinDir,
		debSystemdDir,
	}

	for _, dir := range requiredDebDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	_, err := copy(goBuildBinary, filepath.Join(debBuildDir, commonBinFile))
	if err != nil {
		fmt.Printf("copy from src %s to dest %s failed due to %s\n", goBuildBinary, filepath.Join(debBuildDir, commonBinFile), err)
		os.Exit(1)
	}
	_, err = copy(packageDir+commonSystemdSvcFile, filepath.Join(debBuildDir, commonSystemdDir+commonSystemdSvcFile))
	if err != nil {
		fmt.Printf("copy from src %s to dest %s failed due to %s\n", packageDir+commonSystemdSvcFile, filepath.Join(debBuildDir, commonSystemdDir+commonSystemdSvcFile), err)
		os.Exit(1)
	}

	debTemplates := map[string]string{
		debControlPath:  controlTemplatePath,
		debPostinstPath: postinstTemplatePath,
		debPrermPath:    prermTemplatePath,
	}

	for dest, src := range debTemplates {
		content, err := readFileContent(src)
		if err != nil {
			fmt.Println("Error reading template:", err)
			os.Exit(1)
		}
		processed, err := processTemplate(content, debPackageInfo)
		if err != nil {
			fmt.Println("Error processing template:", err)
			os.Exit(1)
		}
		if err := writeToFile(dest, processed); err != nil {
			fmt.Println("Error writing file:", err)
			os.Exit(1)
		}
	}

	for _, script := range []string{debPostinstPath, debPrermPath} {
		if err := os.Chmod(script, 0755); err != nil {
			fmt.Println("Error setting executable permissions:", err)
			os.Exit(1)
		}
	}

	debOutput := fmt.Sprintf("./linux-installer/packet-sentry-agent_%s_%s.deb", version, arch)
	if err := runCommand("dpkg-deb", "--build", debBuildDir, debOutput); err != nil {
		fmt.Println("Error building DEB package:", err)
		os.Exit(1)
	}
	fmt.Println("DEB package built:", debOutput)
}
