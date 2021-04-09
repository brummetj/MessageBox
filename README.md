# Message Box application

## Overview

This application app is part of a larger social platform, and it provides users with
the ability to send email-like messages back and forth to each other. Each
user has a “mailbox” with the messages that have been sent to them, and they can reply to
any of these messages.

This application is built with GO version 1.15

## Development

This application is broke into 3 different layers. 

1. api layer
2. service layer
3. data layer

The service layer allows the portability of this application. It is the building 
fundamentals and keeps the separation between protocol, data logic, and user interface.
It is the layer that configures the application, and ties the data to the view.

The api layer is any sort of frontend to a user. Whether it exposes endpoints or views,
the idea is to add the api resources you need to work with the service layer that handles the data. It 
has api documentation, and runs the server.

The data layer handles any application logic. It has direct access to the data store and produces any 
response needed for the service layer. In this case we just propagate the data/errors
to our service layer which sends it back to the resource layer.

This application supports custom logger creation, and configuration via yaml files. 
All custom values are available in config/serviceConfig.yml

There are a couple other serviceConfig files which I'll talk about in the next section.
In production these config files should RARELY be touched.

This application uses Couchbase. Couchbase is my favorite for multiple reasons.
1. Its a IN memory key/value json store. Think mongoDB had a baby with redis!
2. Its wicked fast.
3. It provides SQL like queries! Amazing. All those features of sql at your fingertips for a key/value store!.

The application main lives in cmd/http/main.go 

## Deployment

All application management should be used via `scripts/deploy-app.sh`.
Simply run `scripts/deploy-app.sh -h` to see the commands available to you.

Once this source code is cloned, go into `./scripts` here are 3 scripts that are presented.

1. `deploy-app.sh`: it builds the app into docker, runs it, and allows for environment switching.
2. `deploy-db.sh`: this is a simple script to launch the database to start. It can be launched from the `deploy-app.sh` script
3. `configure-db.sh`: This script is not used by any local terminal, it is rather a script for the database to be configured on launch.
    Since this is a demo application i can't have you clicking through the couchbase UI to setup a bucket :)
   

Deployment is managed on a zone level. This idea is taken from a `green/blue` deployment strategy that allows for quick
switches of code bases when new code is introduced. It's similar to a `staging` but it has less downtime because you can quickly
just switch to a different zone once that zone is fully tested and fledged out. This script handles both zones and can be
configured via parameter. There are two config files that directly relate to putting each of this application in their designated
zone environment `config/serviceConfigBlue.yml` and `config/serviceConfigGreen.yml`

The application is proxied and behind a nginx docker server in a docker network. Kubernetes is too much 
configuration with local VMs or production kubernetes clusters, so this was an easier route to demo the challenge. 

## How to RUN! 

On your first run I recommend running the deploy script with these parameters.
`./scripts/deploy-app.sh -b -db` This will run through all application building processes, database launching, then launch the application.
If you're going to run the docker images alone, you will have a bad time! The database configure has a long sleep to prevent any issues, but if for whatever reason 
you see in the logs the app failed to connect, please tear down the docker for the database and relaunch via `deploy-db.sh` scipt, you should see from the docker logs of that image several success messages to know its up and running successfully. 

When everything is built, go to `http://localhost:3001/swagger/index.html` , you will see the swagger page that is presented to the API.
The application won't be ready right away once the script is finished, nginx usually takes a second to be good.

The port is tied to the postman API requests, so it is running on 3001 (the proxy is)
All application logs will be found at `/tmp/MessageBox.log`

Now open the postman tests and request away! 

## how to STOP

you can run a docker ps and see the the docker images running. You can manually remove everything if you please.
Otherwise you can run the `deploy-app.sh` script to kill it. Usually I would break this out to another script but it was nice to stop the images
when re-running when i wanted or if you want to stop a zone and start a new one.

Run `deploy-app.sh -s all -q` to stop the containers and quit! I leave the database alive... for special reason. We shouldn't be 
removing this ever!

To remove the database run `docker stop MessageBoxDB` and `docker rm MessageBoxDB`
