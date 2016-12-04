# An automatic drivers logbook-web-app

For more detailed information about how it works & looks - please have a look at the **ODL.pdf documentation file**!

We are working on a test-server so you can try it out, but have no ETA for this right now.

# License 
This work is licensed under a [![License](https://i.creativecommons.org/l/by-nc-sa/4.0/80x15.png) Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License](https://creativecommons.org/licenses/by-nc-sa/4.0/).
To view a copy of this license, visit http://creativecommons.org/licenses/by-nc-sa/4.0/ or send a letter to Creative Commons, PO Box 1866, Mountain View, CA 94042, USA.

## GODOCS
 are available at [https://godoc.org/?q=https%3A%2F%2Fgithub.com%2FOpenDriversLog%2Fgoodl](https://godoc.org/?q=https%3A%2F%2Fgithub.com%2FOpenDriversLog%2Fgoodl) )

## Necessary changes

Sorry - the time is running out, there are some configurations to be made inside the code instead of central config files.
Please make a global search for "YourServer" and "127.0.0.1" and check if you need to set up any addresses.

# Makefile & Dockerfile magic


### 1st time setup

1. `git clone ssh://git@github.com:OpenDriversLog/goodl.git`
2. the Repositories *webfw*, *goodl-lib*, *goi18n* & *redistore* must be located next to /goodl 
    * **USE THE NEW MAKE TARGET** `make update-deps` *without sudo*
    * this does something similar to:
    * `cd ..`
    * `git clone ssh://git@github.com:OpenDriversLog/webfw.git`
    * `git clone ssh://git@github.com:OpenDriversLog/goodl-lib.git`
    * `git clone ssh://git@github.com:OpenDriversLog/redistore.git`
    * `git clone ssh://git@github.com:OpenDriversLog/goi18n.git`
3. Create an empty directory "goodl-databases" with an empty "databases"-directory in it, also next to /goodl.
4. The dockerimage needs several files, please create a folder "DONTADDTOGIT" with the files from "example-DONTADDTOGIT" and try to fill as much files as possible.
    - deploy_private - a private key for deploying the docker-image
    - gitlab.txt - Access token for an automatic account creating gitlab issues if a user submits an error report
    - mail.txt - Password for your mailaddress to send error-mails if a user submits an error report
    - client_secret.json -> Access secret for Google API https://console.developers.google.com/project/opendriverslog-web-app/apiui/credential for synching address books
    - limesurvey.txt -> Password for limesurvey admin, if you want to create limesurvey invites in E-Mails.
5. `cd goodl`

## Work Locally (after 1st time setup)

using the _Makefile_:

#### Optional / when needed
1. checkout all  __develop__ branches of the repositorys listed under 1st time setup `make update-deps` 
1.5 - when need - `sudo make start-redis` You only need to start redis once!
2. build the docker-image `make build-dev`


#### always needed
1. `sudo make run-dev`
1. Visit http://localhost:4000/alpha/de/odl for the App. Current 
1. `sudo make dev-remove` stops & removes *goodl-dev* container. *goodl_redis* keeps spinning!


## Important Notices!

### Golang dependencies

Are now managed using the [Glide](https://github.com/Masterminds/glide) tool.
Whenever you need a new golang-library __DO NOT__ _NEVER_ use `go get` but instead use

> glide get github.com/whatever/youneed

from inside the `goodl/` directory __inside the container__.

or 

> mv ./vendor/ ./vendor_bak/ 

and wait till all libs get fetched. 

They won't be commited to the repo, but will reside in on your machine (unless your delete the folder manually).

### About the Docker images & containers

There are 3 different Dockerfiles & init-files in the */dockerize* subfolder.
Using the `dev` container, starts a **Go CompileDaemon** so you don't need to restart the container every time you make changes to the go-code. The CompileDaemon rebuilds the goodl-project using `go build /path/to/goodl` whenever it see's change on `.go` and `.c` files. This should also rebuild the dependencies *webfw* & *goodl-lib* since they are mounted into the `vendor/` subfolder of _goodl_.

The _Dockerfile.testing_ checks out the **develop** branch of each of the 3 main Repos.
_Dockerfile.prod_ shoult only be used by __Jenkins__ to deploy to production.

## Testing

### Unit & Integration tests

jup, we got some too...

### Acceptance Tests

Use [agouti](http://agouti.org/)
A Golang acceptance testing framework.
complemented by the Ginkgo BDD testing framework and Gomega matcher library

Acceptance tests (browser-gui tests) and its test-suite reside in

> /tests/acceptance

### Running Tests

#### On Your Dev Machine

_Usually, you don't want to run acceptance tests on your dev machine. Normally you just test it by hand. The only exception I see, is when you actually write those acceptence tests. In this case you most likely want to start them in your dev container manually._

1. `sudo make run-tests` to spawn your Selenium Hub with a *node_chrome* and *node_firefox* alongside
2. `sudo make dev-test-compose-ff` and/or `sudo make dev-test-compose-chrome` in different consoles to run the acceptence test suite
2.5 wait for tests to complete (takes about 10 minutes)
3. (optional) The current application can be found at `localhost:4004/test/_[lang]_/odl/login


#### On -- the dead server --

Jenkins did that on _goodl_ Merge Requests - now you need to do it manually again 

## run on your server

1. `git clone ssh://git@github.com:OpenDriversLog/goodl.git` or `git pull origin/master` to get the latest changes.
1. `sudo make build-server`
2. `sudo make remove-server`
2. `sudo make run-server`
3. open http://yourServer/alpha

### -- the dead server -- run configs
how they are used by jenkins

1. `master` branch for internal use & testing

    @docker run -d \
        --name goodl_intern \
        --link goodl_redis:redis \
        -e ENVIRONMENT=intern \
        -e SMTP_HOST=mail.opendriverslog.de \
        -e SMTP_PORT=587 \
        -v /srv/landingpage2/db/website-dev.sqlite3:/srv/landingpage2/db/website-dev.sqlite3 \
        -v /srv/goodl-alpha/databases-intern:/databases \
        -p 127.0.0.1:9004:4000 \
    odl_go/goodl:latest


2. `development` branch, internal, most likely unstable, newly started after each sucessfully tested & merged Merge Request on GoODL

    docker run -d \
        --name deploy_intern_dev \
        -e ENVIRONMENT=dev-server \
        -e SMTP_HOST=mail.opendriverslog.de \
        -e SMTP_PORT=587 \
        --link jenkins_goodl_redis:redis \
        -v /srv/landingpage2/db/website-dev.sqlite3:/srv/landingpage2/db/website-dev.sqlite3 \
        -v /srv/goodl-alpha/databases-intern:/databases \
        -p 127.0.0.1:9003:4000 \
    odl_jenkins/goodl_mr:live
