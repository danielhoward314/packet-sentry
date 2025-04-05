package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// common constants
const (
	// inputs
	commonLinuxInstallerDir      = "./linux-installer"
	commonBuildDir               = commonLinuxInstallerDir + "/build"
	commonTemplatesDir           = commonLinuxInstallerDir + "/common-templates"
	commonSystemdServiceTemplate = commonTemplatesDir + "/packet-sentry-agent.service.tmpl"

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
	debTemplatesPath        = commonLinuxInstallerDir + "/deb-templates"
	debControlTemplatePath  = debTemplatesPath + "/control.tmpl"
	debPostinstTemplatePath = debTemplatesPath + "/postinst.tmpl"
	debPrermTemplatePath    = debTemplatesPath + "/prerm.tmpl"

	// outputs
	debBuildDir       = commonBuildDir + "/deb"
	debDebianDir      = debBuildDir + "/DEBIAN"
	debControlPath    = debBuildDir + "/DEBIAN/control"
	debPostinstPath   = debBuildDir + "/DEBIAN/postinst"
	debPrermPath      = debBuildDir + "/DEBIAN/prerm"
	debInstallDir     = debBuildDir + commonInstallDir
	debSystemdDir     = debBuildDir + commonSystemdDir
	debSystemdSvcFile = debBuildDir + commonSystemdSvcFile
	debBinDir         = debBuildDir + commonBinDir
	debFinalOut       = commonBuildDir + "/debfinalout"
)

// rpm constants
const (
	// inputs
	rpmTemplatesDir     = commonLinuxInstallerDir + "/rpm-templates"
	rpmSpecTemplatePath = rpmTemplatesDir + "/packet-sentry-agent.spec.tmpl"
	rpmSetupFileName    = "setup.sh"
	rpmSetupFile        = commonLinuxInstallerDir + "/" + rpmSetupFileName

	// ouputs
	rpmBuildDir              = commonBuildDir + "/rpm"
	rpmSourcesDir            = rpmBuildDir + "/SOURCES"
	rpmSpecsDir              = rpmBuildDir + "/SPECS"
	rpmSourcesBinaryTemplate = rpmSourcesDir + "/packet_sentry_linux_" // expects <amd64|arm64> appended
	rpmSourcesSetupFile      = rpmSourcesDir + "/setup.sh"
	rpmSourcesServiceFile    = rpmSourcesDir + "/packet-sentry-agent.service"
	rpmSpecsMainSpec         = rpmSpecsDir + "/packet-sentry-agent.spec"
	rpmFinalOut              = commonBuildDir + "/rpmfinalout"
)

type CommonPackageInfo struct {
	Arch                   string
	BootstrapFile          string
	Name                   string
	SystemdServiceFilePath string
	Version                string
}

type DebPackageInfo struct {
	CommonPackageInfo
	BinFile         string
	InstallDir      string
	MaintainerEmail string
}

type RPMSpecData struct {
	CommonPackageInfo
	BinaryDestDir    string
	BinarySourceName string
	GoArch           string
	ServiceDestDir   string
	SetupScriptDest  string
	SetupScriptName  string
}

type FileWithMode struct {
	File string
	Mode os.FileMode
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

func currentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	return dir
}

func makeRequiredDirs(dirs []FileWithMode) error {
	for _, dir := range dirs {
		fmt.Printf("Making required directory %s with permission %s\n", dir.File, dir.Mode.String())
		if err := os.MkdirAll(dir.File, dir.Mode); err != nil {
			fmt.Printf("Error creating directory %s with permission %s due to: %s\n", dir.File, dir.Mode.String(), err)
			return err
		}
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

func readFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filePath, err)
	}
	return string(content), nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func writeToFile(filePath, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}
	return nil
}

func buildDebPackage(goBuildBinary, version, arch string) {
	fmt.Println("Setting up the .deb build directories")
	requiredDebDirs := []FileWithMode{
		{File: debBuildDir, Mode: 0755},
		{File: debDebianDir, Mode: 0755},
		{File: debInstallDir, Mode: 0755},
		{File: debBinDir, Mode: 0755},
		{File: debSystemdDir, Mode: 0755},
		{File: debFinalOut, Mode: 0755},
	}

	err := makeRequiredDirs(requiredDebDirs)
	if err != nil {
		log.Fatalf("exiting due to: %s\n", err)
	}

	fmt.Println("Copying executable source files into expected deb build directories")
	fmt.Printf("Copying file %s to %s\n", goBuildBinary, filepath.Join(debBuildDir, commonBinFile))
	_, err = copy(goBuildBinary, filepath.Join(debBuildDir, commonBinFile))
	if err != nil {
		log.Fatalf("copy from src %s to dest %s failed due to %s\n", goBuildBinary, filepath.Join(debBuildDir, commonBinFile), err)
	}
	if err := os.Chmod(filepath.Join(debBuildDir, commonBinFile), 0755); err != nil {
		log.Fatalf("Error setting permissions 0755 on file %s due to: %s", filepath.Join(debBuildDir, commonBinFile), err)
	}

	fmt.Println("Parsing and executing templates; copying to expected deb build directories")
	debPackageInfo := DebPackageInfo{
		CommonPackageInfo: CommonPackageInfo{
			Arch:                   arch,
			BootstrapFile:          commonBootstrapFile,
			Name:                   commonPackageName,
			SystemdServiceFilePath: commonSystemdSvcFile,
			Version:                version,
		},
		BinFile:         commonBinFile,
		InstallDir:      commonInstallDir,
		MaintainerEmail: "maintainer@example.com",
	}

	debTemplates := map[string]FileWithMode{
		debControlPath:    {File: debControlTemplatePath, Mode: 0644},
		debPostinstPath:   {File: debPostinstTemplatePath, Mode: 0755},
		debPrermPath:      {File: debPrermTemplatePath, Mode: 0755},
		debSystemdSvcFile: {File: commonSystemdServiceTemplate, Mode: 0644},
	}

	for dest, src := range debTemplates {
		fmt.Printf("Processing .deb template %s for destination %s\n", src.File, dest)
		content, err := readFileContent(src.File)
		if err != nil {
			log.Fatalf("Error reading template %s due to: %s", src.File, err)
		}
		processed, err := processTemplate(content, debPackageInfo)
		if err != nil {
			log.Fatalf("Error processing template %s due to: %s", src.File, err)
		}
		if err := writeToFile(dest, processed); err != nil {
			log.Fatalf("Error writing processed template to file %s due to: %s", dest, err)
		}
		if err := os.Chmod(dest, src.Mode); err != nil {
			log.Fatalf("Error setting permissions %s on file %s due to: %s", src.Mode.String(), dest, err)
		}
	}

	debOutput := fmt.Sprintf(debFinalOut+"/packet-sentry-agent_%s_%s.deb", version, arch)
	fmt.Printf("About to execute dpkg-deb --build %s %s\n", debBuildDir, debOutput)
	if err := runCommand("dpkg-deb", "--build", debBuildDir, debOutput); err != nil {
		log.Fatalf("Error building DEB package: %s", err)
	}
	fmt.Println("DEB package built:", debOutput)
}

func buildRPMPackage(goBuildBinary, version, arch string) {
	fmt.Println("Setting up the RPM build directories")
	requiredRPMDirs := []FileWithMode{
		{File: rpmBuildDir, Mode: 0755},
		{File: rpmSourcesDir, Mode: 0755},
		{File: rpmSpecsDir, Mode: 0755},
	}

	err := makeRequiredDirs(requiredRPMDirs)
	if err != nil {
		log.Fatalf("exiting due to: %s\n", err)
	}

	fmt.Println("Copying executable source files into expected rpm build directory (i.e. SOURCES)")
	sourcesBinaryDest := rpmSourcesBinaryTemplate + arch

	rpmSources := map[string]FileWithMode{
		sourcesBinaryDest:     {File: goBuildBinary, Mode: 0755},
		rpmSourcesServiceFile: {File: commonSystemdServiceTemplate, Mode: 0755},
		rpmSourcesSetupFile:   {File: rpmSetupFile, Mode: 0755},
	}

	for dest, src := range rpmSources {
		_, err := copy(
			src.File,
			dest,
		)
		if err != nil {
			log.Fatalf("copy from src %s to dest %s failed due to %s\n", src.File, dest, err)
		}
		if err := os.Chmod(dest, src.Mode); err != nil {
			log.Fatalf("Error setting permissions %s on file %s due to: %s", src.Mode.String(), dest, err)
		}
	}

	fmt.Println("Parsing and executing templates; copying to expected rpm build directories")
	// Map Go arch to RPM arch
	rpmArch := arch
	if arch == "amd64" {
		rpmArch = "x86_64"
	} else if arch == "arm64" {
		rpmArch = "aarch64"
	}

	rpmPackageInfo := RPMSpecData{
		CommonPackageInfo: CommonPackageInfo{
			Arch:                   rpmArch,
			BootstrapFile:          commonBootstrapFile,
			Name:                   commonPackageName,
			SystemdServiceFilePath: commonSystemdSvcFile,
			Version:                version,
		},
		BinaryDestDir:    commonBinDir,
		BinarySourceName: fmt.Sprintf("packet_sentry_linux_%s", arch),
		GoArch:           arch,
		ServiceDestDir:   commonSystemdDir,
		SetupScriptDest:  commonInstallDir,
		SetupScriptName:  rpmSetupFileName,
	}

	rpmTemplates := map[string]FileWithMode{
		rpmSpecsMainSpec:      {File: rpmSpecTemplatePath, Mode: 0644},
		rpmSourcesServiceFile: {File: commonSystemdServiceTemplate, Mode: 0644},
	}

	for dest, src := range rpmTemplates {
		fmt.Printf("Processing .rpm template %s for destination %s\n", src.File, dest)
		content, err := readFileContent(src.File)
		if err != nil {
			log.Fatalf("Error reading template %s due to: %s", src.File, err)
		}
		processed, err := processTemplate(content, rpmPackageInfo)
		if err != nil {
			log.Fatalf("Error processing template %s due to: %s", src.File, err)
		}
		if err := writeToFile(dest, processed); err != nil {
			log.Fatalf("Error writing processed template to file %s due to: %s", dest, err)
		}
		if err := os.Chmod(dest, src.Mode); err != nil {
			log.Fatalf("Error setting permissions %s on file %s due to: %s", src.Mode.String(), dest, err)
		}
	}

	fmt.Printf(
		"About to execute rpmbuild --define %s --bb %s\n",
		fmt.Sprintf("_topdir %s", filepath.Join(currentDir(), rpmBuildDir)),
		filepath.Join(rpmBuildDir, "SPECS", "packet-sentry-agent.spec"),
	)
	cmd := exec.Command(
		"rpmbuild",
		"--define", fmt.Sprintf("_topdir %s", filepath.Join(currentDir(), rpmBuildDir)),
		"-bb",
		filepath.Join(rpmBuildDir, "SPECS", "packet-sentry-agent.spec"),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to build RPM: %s", err)
	}

	rpmFile := fmt.Sprintf("packet-sentry-agent-*.%s.rpm", rpmArch)
	if err := os.MkdirAll(rpmFinalOut, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %s", err)
	}

	findCmd := exec.Command(
		"find",
		filepath.Join(currentDir(), rpmBuildDir, "RPMS", rpmArch),
		"-name", rpmFile,
		"-type", "f",
	)

	output, err := findCmd.Output()
	if err != nil {
		log.Fatalf("Failed to find built RPM: %v", err)
	}

	rpmPath := strings.TrimSpace(string(output))
	if rpmPath == "" {
		log.Fatalf("Could not find built RPM file matching %s", rpmFile)
	}

	_, rpmFilename := filepath.Split(rpmPath)
	fmt.Printf("Copying file %s to %s\n", rpmPath, filepath.Join(rpmFinalOut, rpmFilename))
	_, err = copy(
		rpmPath,
		filepath.Join(rpmFinalOut, rpmFilename),
	)
	if err != nil {
		log.Fatalf(
			"copy from src %s to dest %s failed due to %s\n",
			rpmPath,
			filepath.Join(rpmFinalOut, rpmFilename),
			err,
		)
	}

	fmt.Printf("RPM file copied to %s\n", filepath.Join(rpmFinalOut, rpmFilename))

	fmt.Printf("RPM build complete for %s architecture\n", arch)
}
func main() {
	if len(os.Args) != 4 {
		log.Fatalf("Invalid number of arguments.\n\n\tUsage: go run main.go [version] [arch] [format]\n\n")
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
	fmt.Printf("Checking for existing go build binary %s\n", goBuildBinary)
	if _, err := os.Stat(goBuildBinary); os.IsNotExist(err) {
		log.Fatalf("Error: Binary %s not found. Run `./scripts/build linux %s` to build it.\n", goBuildBinary, arch)
	}

	switch format {
	case "deb":
		buildDebPackage(goBuildBinary, version, arch)
	case "rpm":
		buildRPMPackage(goBuildBinary, version, arch)
	default:
		log.Fatalf("Unsupported installer format %s\n", format)
	}
}
