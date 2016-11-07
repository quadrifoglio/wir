# wir - Autonomous virtualization system

## Requirements

### Global

* nbd kernel module
* partx (linux-util)
* parted
* e2fsck and resize2fs
* ebtables
* iproute2 (ip command with bridge and tuntap)

### If using QEMU/KVM

* kvm kernel module
* qemu-system
* qemu-kvm
* qemu-img
* qemu-nbd
