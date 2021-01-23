# Landscaper with Dependencies

## Motivation 

Proof on Concept for [[RFC] Dependency Management](https://github.com/gardener/landscaper/issues/97)

## Targets

An installation might only be installed once into a target. Therefore, in kubernetes the target includes the namespace.

## Open topics

* For the special case of the environment broker, we need to create one namespace per environment. See [targets](#targets).
* Finding consensus on different parameter sets
* `component_descriptor` entry point
