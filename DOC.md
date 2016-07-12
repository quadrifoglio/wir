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
[image1, image2...]
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
