# envoy-wrapper

Just a simple wrapper to help envoy shutdown gracefully.  
I copied from [contour envoy shutdown](https://github.com/projectcontour/contour/blob/master/design/envoy-shutdown.md) and tweak a little for my need.

## How it work
envoy-wrapper is designed to catch these following signal `SIGINT`, `SIGTERM`, `SIGHUP`, `SIGQUIT` to handle graceful shutdown for envoy.  
  
It also exposes a webhook `localhost:8090/shutdown` to handle graceful shutdown. 