### v1.6.1
* add a workaround to set non-existent in config values by env vars
### v1.6.0
* Add `-cleanup` flag to clean up the non-cleaned docker staff containerIds and imageIds from previous build  
  non-cleaned docker staff from previous build saves in `isp-test-docker-session_XXX`, where XXX is unix timestamp
* Add new Wait function for connection to config-service
### v1.5.6
* moved event entities from isp-lib to event-lib
### v1.5.5
* add protection for docker Cleanup from stopping tests by the interrupt or kill signals
### v1.5.4
* fix ip in container lifecycle
### v1.5.3
* fix ip in container lifecycle
### v1.5.2
* update libs
### v1.5.0
* Migrated to go modules  
  Warning: this and further releases are incompatible with golang/dep
### v1.4.1
* fix log duplication
### v1.4.0
* attach container logger when run ctx.StartContainer()
### v1.3.0
* add methods to stop and start a container
### v1.2.0
* add method to get bridge address
### v1.1.1
* hotfix
### v1.1.0
* new method to get container address
* new test environment helper
### v1.0.2
* upd pg container
### v1.0.1
* add support docker valumes
### v1.0.0
* package refactoring
### v0.1.0
* init
