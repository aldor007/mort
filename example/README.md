# Mort behind nginx

## Scheme

```apple js
        |==============|
        |  Dark World  |
        |==============|
              ||||
        |==============|
        |   nginx
        |==============|
              | 
              |
         /-----------\
        |    mort    |  
         \-----------/

```
## Assumptions 
* Object will be stored in /var/mort/data
* You will use nginx with slice module
* nginx will terminate SSL
* prometheus job for mort has name "mort"

## Files structure

* config.yml - mort configuration file (copy of demo config)
* deploy-mort.sh - bash script for building and running mort instance
* Dockerfile - simple docker file with configuration for mort
* mort-nginx.config - full nginx configuration that can be easy extanded for any use case
* monitoring.json - example monitoring dashboard 

## Grafana dashboard

<a href="https://mort.mkaciuba.com/demo/dashboard.png"><img src="https://mort.mkaciuba.com/demo/medium/dashboard.png" width="500px"/></a>

