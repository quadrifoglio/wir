# wir

/!\ WORK IN PROGRESS /!\

Self-hosted virtualization platform.

![screenshot](misc/screenshot.png)

## Features

* Create, and manage virtual machines easily
* Single HTTP/JSON API for all the backends
* Web & command line clients

## Backends

* qemu/kvm
* lxc (using liblxc)
* openvz (vzctl)

## Requirements

* one of the supported backends
* bridge-utils
* for qemu sysprep (optional): nbd kernel module, qemu-nbd, partx (util-linux)
