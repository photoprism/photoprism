# photoprism via podman and systemd

You are going to run the containers in *user mode* with `podman`. You would need to set up a (new) user. We assume the username is **photoprism** (otherwise you have to adjust a few files and commands), and that this user is *dedicated* for the task at hand: running the service and hosting the image files. *So we are going to put all directories and files inside this user's home directory*.

## 

## prerequisites

- A linux system
  - using **systemd** as their *init system*
  - having **podman** installed
  - having **git** installed

## 

## creating the user

You are going to create the new user photoprism and afterwards assign a password:

```shell
sudo useradd -m photoprism && sudo passwd photoprism
```

## 

## Set up

**All the next steps are done as the newly created user photoprism.**

```shell
su photoprism
```

### 

### preparations

You are going to need 4 directories, 1 for the database container (`database`) and 3 for the webserver container (`originals`, `import`, `storage`)

```shell
mkdir ~/{database,originals,import,storage}
```

You need to tell the system that this user is allowed to keep processes running even after the user logged out:

```shell
loginctl enable-linger photoprism
```

We are going to store the unit files for systemd in *~/.config/systemd/user/*. As this directory might not exist, we are going to create it:

```shell
mkdir -p ~/.config/systemd/user/
```

### 

### cloning this project

```shell
git clone https://github.com/photoprism/photoprism.git
```

navigate into the repository

```shell
cd photoprism
```

### 

### configuring the containers

All relevant files are located in `podman-systemd`

#### database

By default, a **schema** called *photoprism* as well as a **user** called *photoprism* are created in the database (the DBMS **mariadb**). The user's default password is *insecure*.

You can change this by editing the two files `container-mariadb.service` and `container-webserver.service`:

- schema
  - MARIADB_DATABASE=**photoprism** in `container-mariadb.service`
  - PHOTOPRISM_DATABASE_NAME=**photoprism** in `container-webserver.service`
- user
  - MARIADB_USER=**photoprism** in `container-mariadb.service`
  - PHOTOPRISM_DATABASE_USER=**photoprism** in `container-webserver.service`
- schema
  - MARIADB_PASSWORD=**insecure** in `container-mariadb.service`
  - PHOTOPRISM_DATABASE_PASSWORD=**insecure** in `container-webserver.service`

#### local storage / volumes

The 4 directories you created are being referenced in the following places:

- database
  - **~/database/**:/var/lib/mysql:Z in `container-mariadb.service`
- originals, import, storage
  - **~/originals/**:/photoprism/originals:Z
  - **~/import/**:/photoprism/import:Z
  - **~/storage/**:/photoprism/storage:Z
  - in `container-webserver.service`

#### initial admin password for photoprism

You might want to change the admin password in `container-webserver.service`:

- PHOTOPRISM_ADMIN_PASSWORD="**please-change-me**"

### 

### installing the pod and containers as systemd units

```shell
ln \
 podman-systemd/pod-photoprism.service \
 podman-systemd/container-mariadb.service \
 podman-systemd/container-webserver.service \
 ~/.config/systemd/user/
```

Next we are going to tell systemd about the new units

```shell
systemctl --user daemon-reload
```

## Runing photoprism

Now we can start the pod and both containers

```shell
systemctl --user enable --now pod-photoprism.service
```

let's try to access the the webserver on the command line

```shell
curl http://localhost:2342
```

## next steps / TODO

- proxy, e.g. nginx
- auto updates
