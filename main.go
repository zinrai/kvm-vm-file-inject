package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var fileFlag = flag.String("file", "", "Path to the file to be placed on the VM (required)")
	var dirFlag = flag.String("dir", "", "Target directory path on the VM (required)")
	var stdinFlag = flag.Bool("stdin", false, "Read data from standard input (default if neither -stdin nor -source specified)")
	var sourceFlag = flag.String("source", "", "Path to local source file to read data from")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] VM_NAME\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Place data as a file in a KVM virtual machine.\n")
		fmt.Fprintf(os.Stderr, "Note: For safety, files can only be placed on VMs that are in shutoff state.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Copy from standard input\n")
		fmt.Fprintf(os.Stderr, "  echo \"Hello\" | %s -file hello.txt -dir /home/user vm-name\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Copy from local file\n")
		fmt.Fprintf(os.Stderr, "  %s -source /path/to/local/file.txt -file file.txt -dir /home/user vm-name\n", os.Args[0])
	}

	flag.Parse()

	// Check required options
	if *fileFlag == "" || *dirFlag == "" {
		fmt.Fprintf(os.Stderr, "Error: -file and -dir options are required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check if both stdin and source flags are provided
	if *stdinFlag && *sourceFlag != "" {
		fmt.Fprintf(os.Stderr, "Error: -stdin and -source cannot be used together\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check VM name
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Error: VM name must be specified\n")
		flag.Usage()
		os.Exit(1)
	}
	vmName := args[0]

	// Verify that the VM is shut off
	isShutoff, err := isVMShutoff(vmName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if !isShutoff {
		fmt.Fprintf(os.Stderr, "Error: VM '%s' is not shut off. For safety, files can only be placed on VMs that are in shutoff state.\n", vmName)
		os.Exit(1)
	}

	// Get temporary directory
	tempDir := os.TempDir()
	fileName := filepath.Base(*fileFlag)
	tempFilePath := filepath.Join(tempDir, fileName)

	// Create temporary file
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create temporary file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tempFilePath) // Remove temporary file when function exits

	// Determine input source and copy data to temporary file
	if *sourceFlag != "" {
		// Read from source file
		fmt.Fprintf(os.Stderr, "Reading data from file: %s\n", *sourceFlag)

		sourceFile, err := os.Open(*sourceFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to open source file: %v\n", err)
			os.Exit(1)
		}
		defer sourceFile.Close()

		_, err = io.Copy(tempFile, sourceFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to copy data from source file: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Read from standard input (default or when -stdin is specified)
		fmt.Fprintf(os.Stderr, "Reading data from standard input...\n")

		_, err = io.Copy(tempFile, os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to read data from standard input: %v\n", err)
			os.Exit(1)
		}
	}

	err = tempFile.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to close temporary file: %v\n", err)
		os.Exit(1)
	}

	// Set virt-copy-in command options
	copyArgs := []string{"-d", vmName, tempFilePath, *dirFlag}

	fmt.Fprintf(os.Stderr, "Executing command: virt-copy-in %s\n", strings.Join(copyArgs, " "))

	// Execute virt-copy-in command with sudo
	cmd := exec.Command("sudo", append([]string{"virt-copy-in"}, copyArgs...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "virt-copy-in command execution error: %v\n%s\n", err, output)
		os.Exit(1)
	}

	fmt.Printf("Successfully copied file %s to directory %s on VM %s\n",
		*fileFlag, *dirFlag, vmName)
}

// check if VM is shut off
func isVMShutoff(vmName string) (bool, error) {
	cmd := exec.Command("sudo", "virsh", "list", "--state-shutoff", "--name")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("virsh command execution error: %v", err)
	}

	// Get list of VM names from output
	shutoffVMs := strings.Split(strings.TrimSpace(string(output)), "\n")

	// If output is empty, no VMs are shut off
	if len(shutoffVMs) == 1 && shutoffVMs[0] == "" {
		return false, nil
	}

	// Check if specified VM is in the list of shut off VMs
	for _, vm := range shutoffVMs {
		if strings.TrimSpace(vm) == vmName {
			return true, nil
		}
	}

	return false, nil
}
