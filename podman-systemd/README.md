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
su - photoprism
```

### 

### preparations

You are going to need 4 directories, 1 for the database container (`database`) and 3 for the webserver container (`originals`, `import`, `storage`).

(In case you want to create any of these directories elsewhere or want to name them differently, you'll have to adjust the config; detailed instructions below)

```shell
mkdir ~/{database,originals,import,storage}
```

You need to tell the system that this user is allowed to keep processes running even after the user logged out:

```shell
loginctl enable-linger $(whoami)
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

#### creating your config files

```shell
cp podman-systemd/container-photoprism-database-user.template podman-systemd/container-photoprism-database-user.env
cp podman-systemd/container-photoprism-webserver-user.template podman-systemd/container-photoprism-webserver-user.env
```

#### database

By default, a **schema** called *photoprism* as well as a **user** called *photoprism* are created in the database (the DBMS **mariadb**). The user's default password is *insecure*.

You can change this by editing the two files `container-photoprism-database-user.env` and `container-photoprism-webserver-user.env`:

- schema
  - MARIADB_DATABASE=**photoprism** in `container-photoprism-database-user.env`
  - PHOTOPRISM_DATABASE_NAME=**photoprism** in `container-photoprism-webserver-user.env`
- user
  - MARIADB_USER=**photoprism** in `container-photoprism-database-user.env`
  - PHOTOPRISM_DATABASE_USER=**photoprism** in `ccontainer-photoprism-webserver-user.env`
- password
  - MARIADB_PASSWORD=**insecure** in `container-photoprism-database-user.env`
  - PHOTOPRISM_DATABASE_PASSWORD=**insecure** in `container-photoprism-webserver-user.env`

#### local storage / volumes (optional)

In case you decided to persist your data in non-standard directories or you are running photoprism as another user, you would need to create a new `.env` file and adjust it accordingly:

```shell
cp podman-systemd/volumes-photoprism-user.template podman-systemd/volumes-photoprism-user.env
```

#### initial admin password for photoprism

You might want to change the admin password in `container-photoprism-webserver-user.env`:

- PHOTOPRISM_ADMIN_PASSWORD="**please-change-me**"

### 

### installing the pod and containers as systemd units

```shell
ln -s \
 $(pwd)/podman-systemd/*.service \
 $(pwd)/podman-systemd/*.env \
 ~/.config/systemd/user/
```

Next we are going to tell systemd about the new units

```shell
systemctl --user daemon-reload
```

## Running photoprism

Now we can start the pod and both containers

```shell
systemctl --user enable --now pod-photoprism.service
```

let's try to access the webserver on the command line

```shell
curl http://localhost:2342
```

## Auto updates

Next you are going to enable automatic updates of our images. The containers are being created with this flag: `--label "io.containers.autoupdate=image"` (in the `.service` files). Containers with this label which are controlled via systemd can be automatically updated regularly. To do so, we enable a *systemd timer*:

```shell
systemctl --user enable --now podman-auto-update.timer
```



## next steps / TODO

- proxy, e.g. nginx
