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

## Storage options

* simple directory
* zfs

## Requirements

* one of the supported backends (qemu-system, qemu-utils, lxc-dev)
* bridge-utils
* ebtables
* nbd kernel module, qemu-nbd, partx (util-linux), parted
* for lxc live migration (optional): criu
