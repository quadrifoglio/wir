# Documentation

## Général

Les réponses de l'api se présente de la manière suivante:

```json
{
	"Status": "200",
	"Message": "Success",
	"Content": ...
}
```

L'attribut "Content" n'est pas toujours présent, cela dépend de la requête.
En cas d'erreur, "Status" prends la valeur appropriée et "Message" détail l'erreur survenue.

## Ressources

### GET /

Informations sur le serveur

```json
{
	"Name": string,
	"Version": string,
	"Configuration": {
		"NodeID": int,
		"Address": string (ip address),
		"Port": int,
		"BridgeIface": string,
		"EnableKVM": bool,
		"EbtablesCommand": string (path),
		"QemuImgCommand": string (path),
		"QemuNbdCommand": string (path),
		"QemuCommand": string (path),
		"VzctlCommand": string (path),
		"DatabaseFile": string (path),
		"ImagePath": string (path),
		"MachinePath": string (path)
	},
	"Stats": {
		"CPUUsage": float (percent),
		"RAMUsage": int (MiB),
		"RAMTotal": int (MiB),
		"FreeSpace": int (GiB)
	}
}
```

### GET /images

Listes des images

```json
[image1, image2, ...]
```

### GET /images/*name*

Détails sur une image

```json
{
	"Name": string,
	"Type": string,
	"Source": string (path),
	"MainPartition": int,
	"Arch": string,
	"Distro": string,
	"Release": string
}
```

### POST /images

Créer une image

```json
{
	"Name": string,
	"Type": string,
	"Source": string (URL),
	"MainPartition": int,
	"Arch": string,
	"Distro": string,
	"Release": string
}
```

"Source" peut être de la forme:

* file://path (on the server)
* http://host/path
* scp://user@host/path

### DELETE /images/*name*

Supprimer une image

### GET /machines

Liste des machines

```json
[machine1, machine2, ...]
```

### GET /machines/*name*

Détails sur une machines

```json
{
	"Name": string,
	"Index": int,
	"Type": string,
	"Image": string,
	"State": int (0 down, 1 up),
	"Cores": int,
	"Memory": int (MiB),
	"Network": {
		"Mode": string,
		"MAC": string (mac address),
		"IP": string (ip address)
	},
	"Qemu": {
		"PID": int (pid, only if type is qemu)
	},
	"Vz": {
		"CTID": int (only if type is openvz)
	}
}
```

### POST /machines

Créer une machine

```json
{
	"Name": string,
	"Image": string,
	"Cores": int,
	"Memory": int (MiB),
	"Network": {
		"Mode": string,
		"MAC": string (mac address),
		"IP": string (ip address)
	}
}
```

"Image" doit faire référence à une image existante
"Network" est optionel
"Mode" doit être "bridge" (seul mode supporté à l'heure actuelle)
"MAC" doit être au format aa:bb:cc:dd:ee:ff
"IP" doit être au format décimal pointé

### POST /machines/*name*

Changer les infos d'une machine

```json
{
	"Cores": int,
	"Memory": int (MiB),
	"Network": {
		"Mode": string,
		"MAC": string (mac address),
		"IP": string (ip address)
	}
}
```

Même règles qu'au dessus

### SYSPREP /machines/*name*

Change les données internes d'une machine linux (hostname et mot de passe root)

```json
{
	"Hostname": string,
	"RootPasswd": string
}
```

### START /machines/*name*

Démarre la machine

### STATS /machines/*name*

Obtenir des informations (CPU, Mémoire utilisée sur l'hôte) a propos de la machine

```json
{
	"CPU": float (percent),
	"RAMUsed": int (MiB),
	"RAMFree": int (MiB)
}
```

### STOP /machines/*name*

Arrête la machine

### DELETE /machines/*name*

Supprime la machine