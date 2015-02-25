Wharf - ContainerOps Open Source Platform
=============================

![](http://7vzqdz.com1.z0.glb.clouddn.com/wharf.png)

# How To Compile Wharf Application

Clone code into directory `$GOPATH/src/githhub.com/dockercn` and then exec commands:

```
go get -u github.com/astaxie/beego
go get -u github.com/codegangsta/cli
go get -u github.com/siddontang/ledisdb/ledis
go get -u github.com/garyburd/redigo/redis
go get -u github.com/shurcooL/go/github_flavored_markdown
go build
```

# Wharf Runtime Configuration

Please add a runtime config file named `bucket.conf` under `wharf/conf` before starting `wharf` service.

```ini
runmode = dev

enablehttptls = true
httpsport = 443
httpcertfile = cert/wharf.crt
httpkeyfile = cert/wharf.key

[docker]
BasePath = /tmp/registry
StaticPath = files
Endpoints = 127.0.0.1
Version = 0.8.0
Config = prod
Standalone = true
OpenSignup = false
Gravatar = data/gravatar

[ledisdb]
DataDir = /tmp/ledisdb
DB = 8

[log]
FilePath = /tmp/log
FileName = bucket-log

[session]
Provider = ledis
SavePath = /tmp/session
```

* Application run mode must be `dev` or `prod`.
* If you use Nginx as front end, make sure `enablehttptls` is `false`.
* If run with TLS and without Nginx, set `enablehttptls` is `true` and set the file and key file.
* The `BasePath` is where `Docker` and `Rocket` image files are stored.
* `Endpoints` is very important parameter, set the same value as your domain or IP. For example, you run `wharf` with domain `xxx.org`, then `Endpoints` should be `xxx.org`.
* `DataDir` is where `ledis` data is located.
* The `wharf` session provider default is `ledis`, the `Provider` and `SavePath` is session data storage path.


# Nginx Configuration

It's a Nginx config example. You can change **client_max_body_size** what limited upload file size.

```nginx
upstream wharf_upstream {
  server 127.0.0.1:9911;
}

server {
  listen 80;
  server_name xxx.org;
  rewrite  ^/(.*)$  https://xxx.org/$1  permanent;
}

server {
  listen 443;

  server_name xxx.org;

  access_log /var/log/nginx/xxx.log;
  error_log /var/log/nginx/xxx-errror.log;

  ssl on;
  ssl_certificate /etc/nginx/ssl/xxx/x.crt;
  ssl_certificate_key /etc/nginx/ssl/xxx/x.key;

  client_max_body_size 1024m;
  chunked_transfer_encoding on;

  proxy_redirect     off;
  proxy_set_header   X-Real-IP $remote_addr;
  proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
  proxy_set_header   X-Forwarded-Proto $scheme;
  proxy_set_header   Host $http_host;
  proxy_set_header   X-NginX-Proxy true;
  proxy_set_header   Connection "";
  proxy_http_version 1.1;

  location / {
    proxy_pass         http://wharf_upstream;
  }
}
```

# How To Run

Run behind Nginx:

```bash
./wharf web --address 127.0.0.1 --port 9911
```

Run directly:

```bash
./wharf web --address 0.0.0.0 --port 80
```