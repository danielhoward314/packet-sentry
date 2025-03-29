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

// common constants
const (
	// inputs
	commonBuildDir           = "./linux-installer/build"
	commonPackageDir         = "./linux-installer/package"
	commonPackageServiceFile = commonPackageDir + "/packet-sentry-agent.service"

	// outputs
	commonPackageName    = "packet-sentry-agent"
	commonInstallDir     = "/opt/packet-sentry"
	commonBootstrapFile  = commonInstallDir + "/agentBootstrap.json"
	commonBinDir         = commonInstallDir + "/bin"
	commonBinFile        = commonBinDir + "/packet-sentry-agent"
	commonSystemdDir     = "/etc/systemd/system"
	commonSystemdSvcFile = commonSystemdDir + "/packet-sentry-agent.service"
)

// .deb constants
const (
	// inputs
	debTemplatesPath        = "./linux-installer/deb-templates"
	debControlTemplatePath  = debTemplatesPath + "/control.tmpl"
	debPostinstTemplatePath = debTemplatesPath + "/postinst.tmpl"
	debPrermTemplatePath    = debTemplatesPath + "/prerm.tmpl"

	// outputs
	debBuildDir     = commonBuildDir + "/deb"
	debDebianDir    = debBuildDir + "/DEBIAN"
	debControlPath  = debBuildDir + "/DEBIAN/control"
	debPostinstPath = debBuildDir + "/DEBIAN/postinst"
	debPrermPath    = debBuildDir + "/DEBIAN/prerm"
	debInstallDir   = debBuildDir + commonInstallDir
	debSystemdDir   = debBuildDir + commonSystemdDir
	debBinDir       = debBuildDir + commonBinDir
)

type DebPackageInfo struct {
	Arch            string
	Name            string
	Version         string
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

func buildDebPackage(goBuildBinary string, version string, arch string) {
	// DEB package setup
	debPackageInfo := DebPackageInfo{
		Name:            commonPackageName,
		Version:         version,
		BootstrapFile:   commonBootstrapFile,
		InstallDir:      commonInstallDir,
		MaintainerEmail: "maintainer@example.com",
		Arch:            arch,
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
		fmt.Printf("Making required .deb directory %s\n", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	fmt.Printf("Copying file %s to %s\n", goBuildBinary, filepath.Join(debBuildDir, commonBinFile))
	_, err := copy(goBuildBinary, filepath.Join(debBuildDir, commonBinFile))
	if err != nil {
		fmt.Printf("copy from src %s to dest %s failed due to %s\n", goBuildBinary, filepath.Join(debBuildDir, commonBinFile), err)
		os.Exit(1)
	}
	fmt.Printf("Copying file %s to %s\n", commonPackageServiceFile, filepath.Join(debBuildDir, commonSystemdSvcFile))
	_, err = copy(commonPackageServiceFile, filepath.Join(debBuildDir, commonSystemdSvcFile))
	if err != nil {
		fmt.Printf("copy from src %s to dest %s failed due to %s\n", commonPackageServiceFile, filepath.Join(debBuildDir, commonSystemdSvcFile), err)
		os.Exit(1)
	}

	debTemplates := map[string]string{
		debControlPath:  debControlTemplatePath,
		debPostinstPath: debPostinstTemplatePath,
		debPrermPath:    debPrermTemplatePath,
	}

	for dest, src := range debTemplates {
		fmt.Printf("Processing .deb template %s for destination %s\n", src, dest)
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
		fmt.Printf("Changing file permissions to 0755 for script %s\n", script)
		if err := os.Chmod(script, 0755); err != nil {
			fmt.Println("Error setting executable permissions:", err)
			os.Exit(1)
		}
	}

	debOutput := fmt.Sprintf("./linux-installer/packet-sentry-agent_%s_%s.deb", version, arch)
	fmt.Printf("About to execute dpkg-deb --build %s %s\n", debBuildDir, debOutput)
	if err := runCommand("dpkg-deb", "--build", debBuildDir, debOutput); err != nil {
		fmt.Println("Error building DEB package:", err)
		os.Exit(1)
	}
	fmt.Println("DEB package built:", debOutput)
}

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Invalid number of arguments.\n\n\tUsage: go run main.go [version] [arch] [format]\n\n")
		os.Exit(1)
	}

	version := os.Args[1]
	arch := os.Args[2]
	format := os.Args[3]
	fmt.Printf(
		"Building linux installer for version %s, architecture %s and installer format %s\n",
		version,
		arch,
		format,
	)

	goBuildBinary := fmt.Sprintf("./build/packet_sentry_linux_%s", arch)
	fmt.Printf("Checking for existing of go build binary %s\n", goBuildBinary)
	if _, err := os.Stat(goBuildBinary); os.IsNotExist(err) {
		fmt.Printf("Error: Binary %s not found. Run `./scripts/build linux %s` to build it.\n", goBuildBinary, arch)
		os.Exit(1)
	}

	switch format {
	case "deb":
		buildDebPackage(goBuildBinary, version, arch)
	}
}
