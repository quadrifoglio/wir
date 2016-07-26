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
* lxc
* openvz

## Requirements

* one of the supported backends (qemu-system, qemu-utils, lxc-dev, vzctl)
* bridge-utils
* ebtables
* for qemu sysprep (optional): nbd kernel module (max_part=12), qemu-nbd, partx (util-linux)
