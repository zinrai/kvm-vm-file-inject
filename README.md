# kvm-vm-file-inject

A command-line tool to inject files into KVM virtual machines.

## Overview

`kvm-vm-file-inject` allows you to place files into a KVM virtual machine's filesystem from either standard input or a local file.

## Features

- Inject content from standard input or local files
- Safety check to ensure VM is shut off before file injection
- Uses `virt-copy-in` behind the scenes for reliable file operations

## Prerequisites

- `libvirt-clients` package (for `virsh` command)
- `libguestfs-tools` package (for `virt-copy-in` command)
- Sudo privileges (for executing `virsh` and `virt-copy-in`)

## Installation

```bash
$ go build
```

## Usage

```
Usage: kvm-vm-file-inject [options] VM_NAME

Options:
  -dir string
        Target directory path on the VM (required)
  -file string
        Path to the file to be placed on the VM (required)
  -source string
        Path to local source file to read data from
  -stdin
        Read data from standard input (default if neither -stdin nor -source specified)
```

## Examples

Copy from standard input:

```bash
echo "Hello World" | kvm-vm-file-inject -file hello.txt -dir /home/user vm-name
```

Copy from local file:

```bash
$ kvm-vm-file-inject -source /path/to/local/file.txt -file remote-file.txt -dir /home/user vm-name
```

## License

This project is licensed under the [MIT License](./LICENSE).
