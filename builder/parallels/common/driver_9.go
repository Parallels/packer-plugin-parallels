// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ChrisTrenkamp/goxpath"
	"github.com/ChrisTrenkamp/goxpath/tree/xmltree"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
)

// Struct to format keyCodes in JSON
type keyCodes struct {
	KeyCode int    `json:"key"`
	Event   string `json:"event"`
	Delay   int    `json:"delay"`
}

// sending scancodes to VM via prlctl send-key-event CMD
func (d *Parallels9Driver) sendJsonScancodes(vmName string, inputScanCodes []string) error {

	log.Println("scancodes received for JSON encoding ", inputScanCodes)
	delay := 100
	var inputHexScanCodes []uint
	for _, hexStr := range inputScanCodes {
		var val uint
		fmt.Sscanf(hexStr, "%X", &val) // Convert hex string to uint
		inputHexScanCodes = append(inputHexScanCodes, val)
	}
	keyCodeData := []keyCodes{}
	for i := 0; i < len(inputHexScanCodes); i++ {
		key1 := inputHexScanCodes[i]
		if key1 == 224 {
			key2 := inputHexScanCodes[i+1]
			i = i + 1
			if key2 < 128 {
				keyCodeData = append(keyCodeData, keyCodes{KeyCode: getKeycodeFromScanCode([]uint{key1, key2}), Event: "press", Delay: delay})
			} else {
				keyCodeData = append(keyCodeData, keyCodes{KeyCode: getKeycodeFromScanCode([]uint{key1, key2 - 128}), Event: "release", Delay: delay})
			}
		} else if key1 < 128 {
			keyCodeData = append(keyCodeData, keyCodes{KeyCode: getKeycodeFromScanCode([]uint{key1}), Event: "press", Delay: delay})
		} else {
			keyCodeData = append(keyCodeData, keyCodes{KeyCode: getKeycodeFromScanCode([]uint{key1 - 128}), Event: "release", Delay: delay})
		}
	}
	jsonFormat, err := json.MarshalIndent(keyCodeData, "", "\t")
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("complete scancode data in JSON format %s", jsonFormat)

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(d.PrlctlPath, "send-key-event", vmName, "-j")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(string(jsonFormat))
	err = cmd.Run()
	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())
	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("prlctl error: %s", stderrString)
	}
	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Parallels9Driver is a base type for Parallels builders.
type Parallels9Driver struct {
	// This is the path to the "prlctl" application.
	PrlctlPath string

	// This is the path to the "prlsrvctl" application.
	PrlsrvctlPath string

	// The path to the parallels_dhcp_leases file
	dhcpLeaseFile string
}

// Import creates a clone of the source VM and reassigns the MAC address if needed.
func (d *Parallels9Driver) Import(name, srcPath, dstDir string, reassignMAC bool) error {
	err := d.Prlctl("register", srcPath, "--preserve-uuid")
	if err != nil {
		return err
	}

	srcID, err := getVMID(srcPath)
	if err != nil {
		return err
	}

	srcMAC := "auto"
	if !reassignMAC {
		srcMAC, err = getFirstMACAddress(srcPath)
		if err != nil {
			return err
		}
	}

	err = d.Prlctl("clone", srcID, "--name", name, "--dst", dstDir)
	if err != nil {
		return err
	}

	err = d.Prlctl("unregister", srcID)
	if err != nil {
		return err
	}

	err = d.Prlctl("set", name, "--device-set", "net0", "--mac", srcMAC)
	if err != nil {
		return err
	}
	return nil
}

func getVMID(path string) (string, error) {
	return getConfigValueFromXpath(path, "/ParallelsVirtualMachine/Identification/VmUuid")
}

func getFirstMACAddress(path string) (string, error) {
	return getConfigValueFromXpath(path, "/ParallelsVirtualMachine/Hardware/NetworkAdapter[@id='0']/MAC")
}

func getConfigValueFromXpath(path, xpath string) (string, error) {
	file, err := os.Open(path + "/config.pvs")
	if err != nil {
		return "", err
	}

	doc, err := xmltree.ParseXML(file)
	if err != nil {
		return "", err
	}

	xpExec := goxpath.MustParse(xpath)
	node, err := xpExec.Exec(doc)
	if err != nil {
		return "", err
	}

	return node.String(), nil
}

// Finds an application bundle by identifier (for "darwin" platform only)
func getAppPath(bundleID string) (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command("mdfind", "kMDItemCFBundleIdentifier ==", bundleID)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	pathOutput := strings.TrimSpace(stdout.String())
	if pathOutput == "" {
		if fi, err := os.Stat("/Applications/Parallels Desktop.app"); err == nil {
			if fi.IsDir() {
				return "/Applications/Parallels Desktop.app", nil
			}
		}

		return "", fmt.Errorf(
			"Could not detect Parallels Desktop! Make sure it is properly installed.")
	}

	return pathOutput, nil
}

// CompactDisk performs the compaction of the specified virtual disk image.
func (d *Parallels9Driver) CompactDisk(diskPath string) error {
	prlDiskToolPath, err := exec.LookPath("prl_disk_tool")
	if err != nil {
		return err
	}

	// Analyze the disk content and remove unused blocks
	command := []string{
		"compact",
		"--hdd", diskPath,
	}
	if err := exec.Command(prlDiskToolPath, command...).Run(); err != nil {
		return err
	}

	// Remove null blocks
	command = []string{
		"compact", "--buildmap",
		"--hdd", diskPath,
	}
	if err := exec.Command(prlDiskToolPath, command...).Run(); err != nil {
		return err
	}

	return nil
}

// DeviceAddCDROM adds a virtual CDROM device and attaches the specified image.
func (d *Parallels9Driver) DeviceAddCDROM(name string, image string) (string, error) {
	command := []string{
		"set", name,
		"--device-add", "cdrom",
		"--image", image,
		"--enable", "--connect",
	}

	out, err := exec.Command(d.PrlctlPath, command...).Output()
	if err != nil {
		return "", err
	}

	deviceRe := regexp.MustCompile(`\s+(cdrom\d+)\s+`)
	matches := deviceRe.FindStringSubmatch(string(out))
	if matches == nil {
		return "", fmt.Errorf(
			"Could not determine cdrom device name in the output:\n%s", string(out))
	}

	deviceName := matches[1]
	return deviceName, nil
}

// DiskPath returns a full path to the first virtual disk drive.
func (d *Parallels9Driver) DiskPath(name string) (string, error) {
	out, err := exec.Command(d.PrlctlPath, "list", "-i", name).Output()
	if err != nil {
		return "", err
	}

	HDDRe := regexp.MustCompile("hdd0.* image='(.*)' type=*")
	matches := HDDRe.FindStringSubmatch(string(out))
	if matches == nil {
		return "", fmt.Errorf(
			"Could not determine hdd image path in the output:\n%s", string(out))
	}

	HDDPath := matches[1]
	return HDDPath, nil
}

// IsRunning determines whether the VM is running or not.
func (d *Parallels9Driver) IsRunning(name string) (bool, error) {
	var stdout bytes.Buffer

	cmd := exec.Command(d.PrlctlPath, "list", name, "--no-header", "--output", "status")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return false, err
	}

	log.Printf("Checking VM state: %s\n", strings.TrimSpace(stdout.String()))

	for _, line := range strings.Split(stdout.String(), "\n") {
		if line == "running" {
			return true, nil
		}

		if line == "suspended" {
			return true, nil
		}
		if line == "paused" {
			return true, nil
		}
		if line == "stopping" {
			return true, nil
		}
	}

	return false, nil
}

// Stop forcibly stops the VM.
func (d *Parallels9Driver) Stop(name string) error {
	if err := d.Prlctl("stop", name, "--kill"); err != nil {
		return err
	}

	// We sleep here for a little bit to let the session "unlock"
	time.Sleep(2 * time.Second)

	return nil
}

// Prlctl executes the specified "prlctl" command.
func (d *Parallels9Driver) Prlctl(args ...string) error {
	_, err := d.PrlctlGet(args...)
	return err
}

// PrlctlGet executes the given "prlctl" command and returns the output
func (d *Parallels9Driver) PrlctlGet(args ...string) (string, error) {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing prlctl: %#v", args)
	cmd := exec.Command(d.PrlctlPath, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("prlctl error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return stdoutString, err
}

// Verify raises an error if the builder could not be used on that host machine.
func (d *Parallels9Driver) Verify() error {
	return nil
}

// Version returns the version of Parallels Desktop installed on that host.
func (d *Parallels9Driver) Version() (string, error) {
	out, err := exec.Command(d.PrlctlPath, "--version").Output()
	if err != nil {
		return "", err
	}

	versionRe := regexp.MustCompile(`prlctl version (\d+\.\d+.\d+)`)
	matches := versionRe.FindStringSubmatch(string(out))
	if matches == nil {
		return "", fmt.Errorf(
			"Could not find Parallels Desktop version in output:\n%s", string(out))
	}

	version := matches[1]
	return version, nil
}

// SendKeyScanCodes sends the specified scancodes as key events to the VM.
// It is performed using "Prltype" script (refer to "prltype.go") if version is  < 19.0.0.
// scancodes are sent by using prlctl CMD if version is  >= 19.0.0
func (d *Parallels9Driver) SendKeyScanCodes(vmName string, codes ...string) error {
	var stdout, stderr bytes.Buffer
	var err error

	if len(codes) == 0 {
		log.Printf("No scan codes to send")
		return nil
	}

	prlctlCurrVersionStr, verErr := d.Version()
	if verErr != nil {
		return err
	}
	prlctlCurrVersion, _ := version.NewVersion(prlctlCurrVersionStr)
	v2, _ := version.NewVersion("19.0.0")

	if prlctlCurrVersion.LessThan(v2) {

		f, err := tmp.File("prltype")
		if err != nil {
			return err
		}
		defer os.Remove(f.Name())

		script := []byte(Prltype)
		_, err = f.Write(script)
		if err != nil {
			return err
		}

		args := prepend(vmName, codes)
		args = prepend(f.Name(), args)
		cmd := exec.Command("/usr/bin/python3", args...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()

		stdoutString := strings.TrimSpace(stdout.String())
		stderrString := strings.TrimSpace(stderr.String())

		if _, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("prltype error: %s", stderrString)
			return err
		}

		log.Printf("stdout: %s", stdoutString)
		log.Printf("stderr: %s", stderrString)
	} else {
		err = d.sendJsonScancodes(vmName, codes)
	}
	return err
}

func prepend(head string, tail []string) []string {
	tmp := make([]string, len(tail)+1)
	for i := 0; i < len(tail); i++ {
		tmp[i+1] = tail[i]
	}
	tmp[0] = head
	return tmp
}

// SetDefaultConfiguration applies pre-defined default settings to the VM config.
func (d *Parallels9Driver) SetDefaultConfiguration(vmName string) error {
	commands := make([][]string, 5)
	commands[0] = []string{"set", vmName, "--startup-view", "same"}
	commands[1] = []string{"set", vmName, "--on-shutdown", "close"}
	commands[2] = []string{"set", vmName, "--on-window-close", "keep-running"}
	commands[3] = []string{"set", vmName, "--auto-share-camera", "off"}
	commands[4] = []string{"set", vmName, "--smart-guard", "off"}

	for _, command := range commands {
		err := d.Prlctl(command...)
		if err != nil {
			return err
		}
	}
	return nil
}

// MAC returns the MAC address of the VM's first network interface.
func (d *Parallels9Driver) MAC(vmName string) (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command(d.PrlctlPath, "list", "-i", vmName)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		log.Printf("MAC address for NIC: nic0 on Virtual Machine: %s not found!\n", vmName)
		return "", err
	}

	stdoutString := strings.TrimSpace(stdout.String())
	re := regexp.MustCompile("net0.* mac=([0-9A-F]{12}) card=.*")
	macMatch := re.FindAllStringSubmatch(stdoutString, 1)

	if len(macMatch) != 1 {
		return "", fmt.Errorf("MAC address for NIC: nic0 on Virtual Machine: %s not found!\n", vmName)
	}

	mac := macMatch[0][1]
	log.Printf("Found MAC address for NIC: net0 - %s\n", mac)
	return mac, nil
}

// Parses the file /Library/Preferences/Parallels/parallels_dhcp_leases
// file contain a list of DHCP leases given by Parallels Desktop
// Example line:
// 10.211.55.181="1418921112,1800,001c42f593fb,ff42f593fb000100011c25b9ff001c42f593fb"
// IP Address   ="Lease expiry, Lease time, MAC, MAC or DUID"
func (d *Parallels9Driver) ipWithLeases(mac string, vmName string) (string, error) {
	if len(mac) != 12 {
		return "", fmt.Errorf("Not a valid MAC address: %s. It should be exactly 12 digits.", mac)
	}

	leases, err := ioutil.ReadFile(d.dhcpLeaseFile)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile("(.*)=\"(.*),(.*)," + strings.ToLower(mac) + ",.*\"")
	mostRecentIP := ""
	mostRecentLease := uint64(0)
	for _, l := range re.FindAllStringSubmatch(string(leases), -1) {
		ip := l[1]
		expiry, _ := strconv.ParseUint(l[2], 10, 64)
		leaseTime, _ := strconv.ParseUint(l[3], 10, 32)
		log.Printf("Found lease: %s for MAC: %s, expiring at %d, leased for %d s.\n", ip, mac, expiry, leaseTime)
		if mostRecentLease <= expiry-leaseTime {
			mostRecentIP = ip
			mostRecentLease = expiry - leaseTime
		}
	}

	return mostRecentIP, nil
}

func (d *Parallels9Driver) ipWithPrlctl(vmName string) (string, error) {
	ip := ""
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(d.PrlctlPath, "list", vmName, "--full", "--no-header", "-o", "ip_configured")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("Command run failed for Virtual Machine: %s\n", vmName)
		return "", err
	}

	stderrString := strings.TrimSpace(stderr.String())
	if len(stderrString) > 0 {
		log.Printf("stderr: %s", stderrString)
	}

	stdoutString := strings.TrimSpace(stdout.String())
	re := regexp.MustCompile(`([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)`)
	macMatch := re.FindAllStringSubmatch(stdoutString, 1)

	if len(macMatch) != 1 {
		return "", fmt.Errorf("Unable to retrieve ip address of VM %s through tools\n", vmName)
	}

	ip = macMatch[0][1]
	return ip, nil
}

// IPAddress finds the IP address of a VM connected that uses DHCP by its MAC address
// If the MAC address is not found in the DHCP lease file, it will try to find ip using prlctl
func (d *Parallels9Driver) IPAddress(mac string, vmName string) (string, error) {
	ip, err := d.ipWithLeases(mac, vmName)
	if (err != nil) || (len(ip) == 0) {
		log.Printf("IP lease not found for MAC address %s in: %s\n", mac, d.dhcpLeaseFile)

		ip, err = d.ipWithPrlctl(vmName)
		if (err != nil) || (len(ip) == 0) {
			return "", fmt.Errorf("IP address not found for this VM using prlctl: %s\n", vmName)
		}
	}

	log.Printf("Found IP lease: %s for MAC address %s\n", ip, mac)
	return ip, nil
}

// ToolsISOPath returns a full path to the Parallels Tools ISO for the specified guest
// OS type. The following OS types are supported: "win", "lin", "mac", "other".
func (d *Parallels9Driver) ToolsISOPath(k string) (string, error) {
	appPath, err := getAppPath("com.parallels.desktop.console")
	if err != nil {
		return "", err
	}

	toolsPath := filepath.Join(appPath, "Contents", "Resources", "Tools", "prl-tools-"+k+".iso")
	log.Printf("Parallels Tools path: '%s'", toolsPath)
	return toolsPath, nil
}
