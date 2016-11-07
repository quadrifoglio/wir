# wir - API Documentation

## Resources

### Remote

```json
{
	"Host": string (IP Address or hostname)
	"Port": int (TCP port number)
}
```

### Index

```json
{
	"Hostname": string  (System hostname)
	"CpuUsage": float32 (Current CPU Usage in percent)
	"MemoryUsage": uint64  (Currently used memory in KiB)
	"MemoryTotal": uint64  (Total memory available to the system in KiB)
}
```

### Image

```json
{
	"ID": string (64 bit random unique identifier)
	"Name": string (Name of the image)
	"Type": string (Type of the image (kvm, lxc))
	"Source": string (Location of the image file (scheme://[user@]host/path))
}
```

### Network

```json
	"ID" string (64 bit random unique identifier)
	"Name": string (Name of the image)
	"CIDR": string (CIDR notation of network address (a.b.c.d(sk)))
	"GatewayIface": string (Name of a physical interface that should be part of the network (optional))

	"DHCP": {
		"Enabled": bool (Wether internal DHCP is in use on this network)
		"StartIP": string (First IP to be leased)
		"NumIP": int (Number of IP addresses to lease, starting from StartIP)
		"Router": string (IP address of the network router)
	}
}
```

### Volume

```json
{
	"ID": string (64 bit random unique identifier)
	"Name": string (Name of the volume)
	"Type": string (Type of the volume (kvm, lxc))
	"Size": uint64 (Size of the volume in KiB)
}
```

### Machine

```json
{
	"ID": string (64 bit random unique identifier)
	"Name": string (Name of the machine)
	"Image": string (ID of the machine's image)
	"Cores": int (Number of CPUs)
	"Memory": uint64 (Memory in MiB)
	"Disk": uint64 (Disk size in bytes)

	"Volumes": []string (IDs of the attached volumes)
	"Interfaces": [
		{
			"Network": string (ID of the network)
			"MAC": string (MAC address)
			"IP": string (IP address)
		},
		...
	]
}
```

### Machine status

```json
{
	"Running": bool (True if the machine is currently running)
	"CpuUsage": float32 (Percentage of the time the CPU is busy)
	"RamUsage": uint64 (Currently used RAM in MiB)
	"DiskUsage": uint64 (Current size of the disk image in bytes)
}
```

### Checkpoint

```json
{
	"Name": string (Checkpoint name)
	"Timestamp": int64 (Unix timestamp)
}
```

### KVM options

```json
{
	"PID": int (The QEMU/KVM proccess ID)
	"CDRom": string (Path to a disk image to insert into the machine as a CD-ROM)

	"VNC": {
		"Enabled": bool (Wether to use the VNC server)
		"Address": string (Bind address of the VNC server (ip:port))
		"Port": int (Port number)
	}

	"Linux": {
		"Hostname": string (Linux hostname)
		"RootPassword": string (Linux root password in clear text)
	}
}
```

## Endpoints

### /

Resource: Index

* GET / : Get server informations

### /images

Resource: Image

* POST / : Create a new image
* GET  / : List images

* GET    /<id> : Get image information
* POST   /<id> : Update image information
* DELETE /<id> : Delete image

* GET /<id>/data : Get image binary data

### /networks

Resource: Network

* POST / : Create a new network
* GET  / : List networks

* GET    /<id> : Get network information
* POST   /<id> : Update network information
* DELETE /<id> : Delete network

### /volumes

Resource: Volume

* POST / : Create a new volume
* GET  / : List volumes

* GET    /<id> : Get volume information
* POST   /<id> : Update volume information
* DELETE /<id> : Delete volume

### /machines

Resource: Machine

* POST / : Create a new machine
* GET  / : List machines

* POST /fetch : Fetch a virtual machine from a distant node
	* Resource: MachineFetch

* GET    /<id> : Get machine information
* POST   /<id> : Update machine information
* DELETE /<id> : Delete machine

#### Actions

Resource: none

* GET /<id>/start  : Start machine
* GET /<id>/stop   : Stop machine

#### VKM specific options

Resource: KVM options

* GET  /<id>/kvm : Get machine KVM options
* POST /<id>/kvm : Update machine KVM options

#### Other

* GET /<id>/status : Machine status and resource usage
	* Resource: MachineStatus

* GET /disk/data : Main hard drive binary data
	* Resource: None
