# PM

Simple backend for managing your accounts credentials from different services.

## Features

- The application runs in one simple docker container on your system;
- All accounts are stored in encrypted form;
- In case of emergency or scheduled termination of the program, all data is saved to the local storage (volume);
- At the first start, the application will generate the master key for creation backups (the master key will be displayed in the application log also at the start).

## Installation (Linux)

- Create a directory where the application code will be stored and navigate to it:

```shell script
mkdir ~/pm
cd ~/pm
```

- Create a directory where your backups will be stored:

```shell script
mkdir backups
```

- Clone the code from the repository:

```shell script
git clone https://github.com/morinick/pm .
```

- Delete .git directory:

```shell script
rm -rf .git
```

- Run the command to create a docker image (make sure that you already have docker installed):

```shell script
docker build . -t pm-image:latest
```

- Run the docker image with the following command:

```shell script
docker run \
	-p 5000:5000 \
	--name pm \
	-v ./backups:/backup \
	-d \
	pm-image
```

## Start the application if there is a backup

- Using the docker run command (the .env file must store the MASTER_KEY value):

```shell script
docker run \
	-p 5000:5000 \
	--name pm \
	-v ./backups:/backup \
	--env-file /path/to/your/.env
	-d \
	pm-image	
```
