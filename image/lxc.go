package image

var (
	LxcTemplates = []Image{
		Image{"lxc-alpine", TypeLXC, "alpine"},
		Image{"lxc-arch", TypeLXC, "arch"},
		Image{"lxc-centos", TypeLXC, "centos"},
		Image{"lxc-debian", TypeLXC, "debian"},
		Image{"lxc-fedora", TypeLXC, "fedora"},
		Image{"lxc-ubuntu", TypeLXC, "ubuntu"},
	}
)
