Independent algorithm (testing only):

- swirlservice (Golang)

This algorithm can be compiled and run with the settings in defaultconfig.json. This can run for a long time and requires a lot of memory! Results are printed to stdout.

Kubernetes implementation:

- edgeservice (Golang)
- fogservice (Golang)
- swirlservice (Golang)

These folders have a basic Dockerfile to build the services into a container. All can be run in Kubernetes pods, just mind the configs (see defaultconfig.json for the fallback config).

Topology generator will be included soon.
