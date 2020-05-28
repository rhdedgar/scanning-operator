# scanning-operator

## Description
The openshift-scanning-operator manages components related to malware scanning on OpenShift V4. 
There are currently 5 containers in 2 different pods. Each pod is managed by its own DaemonSet.

## Components

### scanner
The scanner pod is made of 4 separate containers which run on all nodes:

#### clamsig-puller
clamsig-puller is responsible for checking the clam signature mirror bucket every 12 hours for new official, unofficial, and custom SRE clamAV signatures.
It then stores those signature databases in a shared volume for use by clamd.

#### container-info
Listens for container IDs from watcher, and returns information about that container in crictl or runc output formats.

#### clamd
The clamAV daemon itself. It receives file descriptors from watcher and does the actual scanning of files.
It loads its signature databases from the shared volume.

#### watcher
Watches the journal for new container creation events. When a new container start event is found, it does the following:

* Gathers information about the new container's pod.
* Queues the container for scanning.
* Sends the scan results to the Logger OpenShift service

### logger

The logger pod has 1 container, which runs in a DaemonSet on master nodes. The basic data flow is:

* Listen for positive scan results and pod creation logs sent by the watcher container. 
* These logs are formatted with additional info about the pod and user from the OpenShift API.
* Pod creation logs and positive scan results are then picked up by the Splunk forwarder.


## Source repos

* https://github.com/openshift/scanning-operator
* https://github.com/openshift/pod-logger
* https://github.com/openshift/pleg-watcher
* https://github.com/openshift/container-info
* https://github.com/openshift/clamsig-puller
* https://github.com/openshift/clamd
* https://github.com/openshift/clam-update
