# VersionCtl
This Go web application queries our internal HaProxy's for available services via [REST HaProxy](https://github.com/malnick/rest_haproxy) then queries all available backend services on their management endpoints for running services. It then compares this to the versions mapped in configuration management (we use puppet, and expose a single hieradata file called versions.yaml, which is managed by Jenkins via a webhook on our Puppet master). 

If the versions running or mapped in configuration management are the same, the service is green; if not, the service is red. This gives us vital visability into the state of a given deployment across all nodes and processes. 
