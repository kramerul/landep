# Landscaper with Dependencies

## Motivation 

Proof on Concept for [[RFC] Dependency Management](https://github.com/gardener/landscaper/issues/97)

## Targets

A `Target` is used to describe the location where an package is installed. 
For packages that are building bridges it's also necessary to have the ability to install them into different targets.
In that case it would be necessary to define something like  `K8sCloudFoundryBridgingTarget`

An installation might only be installed once into a target. Therefore, in kubernetes the target includes the namespace.

## Names

The names of the installed software (e.g helm release) are build using a digest of the target and the installer.

## Conflict resolution of shared installations

See `installation_test.go`

## Open topics

* For the special case of the environment broker, we need to create one namespace per environment. See [targets](#targets).
* `component_descriptor` entry point
