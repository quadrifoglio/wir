# wir

Self-hosted virtualization platform.

![screenshot][misc/screenshot.png]

## Features

* Create, and manage virtual machines easily
* Single HTTP/JSON API for all the backends
* Web & command line clients

## Backends

* qemu/kvm
* lxc (using liblxc-dev)
* openvz (vzctl)

## Requirements

* Go
* bridge-utils
* At lease one of the supported backends
