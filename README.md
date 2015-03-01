# Wharf - ContainerOps Open Source Platform

**This is early stuff. You've been warned.**

Wharf is a successor of ContainerOps platform whose concept is built upon the DevOps, which means it's a higher level solution over traditional approach and DevOps, but not an alternative. The ultimate goal of Wharf is building a pipeline form development to deployment & operations through Docker, Rocket and other container solutions.

ContainerOps is all about product workflow. Wharf builds and runs a container image you defined whenever new code is being pushed, along with corresponding dependence system. But it's not all, the image will also work with continuous integration, or continuous deployment with Rocket, LXC or Atomic, etc. Wharf is focusing on the continuous changes from version control system to production environment.

Now, it's time to announce our pre-release and core of ContainerOps version of Wharf, you can replace Docker Registry with it. Please make sure you know that Wharf is currently under alpha stage.

Our team is still working really hard for your happiness with the complete version of ContainerOps platform, which comes in next few months.

![](http://7vzqdz.com1.z0.glb.clouddn.com/wharf.png)

# How To Compile Wharf Application

Clone code into directory `$GOPATH/src/githhub.com/dockercn` and then exec commands:

```bash
go get -u github.com/astaxie/beego
go get -u github.com/codegangsta/cli
go get -u github.com/siddontang/ledisdb/ledis
go get -u github.com/garyburd/redigo/redis
go get -u github.com/shurcooL/go/github_flavored_markdown
go get -u github.com/satori/go.uuid
go get -u github.com/nfnt/resize
go build
```

# Wharf Runtime Configuration

Please add a runtime config file named `bucket.conf` under `wharf/conf` before starting `wharf` service.

```ini
runmode = dev

enablehttptls = true
httpsport = 443
httpcertfile = cert/containerops.me/containerops.me.crt
httpkeyfile = cert/containerops.me/containerops.me.key

gravatar = data/gravatar

[docker]
BasePath = /tmp/registry
StaticPath = files
Endpoints = containerops.me
Version = 0.8.0
Config = prod
Standalone = true
OpenSignup = false

[ledisdb]
DataDir = /tmp/ledisdb
DB = 8

[log]
FilePath = /tmp
FileName = containerops.log

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
* The bucket.conf should be in folder conf with app.conf. If you wanna change the bucket.conf name, you should be modify the include bucket.conf in theapp.conf last line.

# Nginx Configuration

It's a Nginx config example. You can change **client_max_body_size** what limited upload file size.

You should copy `containerops.me` keys from `cert/containerops.me` to `/etc/nginx`, then run **Wharf** with `http` mode and listen on `127.0.0.1:9911`.

```nginx
upstream wharf_upstream {
  server 127.0.0.1:9911;
}

server {
  listen 80;
  server_name containerops.me;
  rewrite  ^/(.*)$  https://containerops.me/$1  permanent;
}

server {
  listen 443;

  server_name containerops.me;

  access_log /var/log/nginx/containerops-me.log;
  error_log /var/log/nginx/containerops-me-errror.log;

  ssl on;
  ssl_certificate /etc/nginx/containerops.me.crt;
  ssl_certificate_key /etc/nginx/containerops.me.key;

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

# Reporting Issues

Please submit issue at https://github.com/dockercn/wharf/issues

# Maintainers

* Meaglith Ma https://twitter.com/genedna
* Allen Chen https://github.com/chliang2030598
* Leo Meng https://github.com/fivestarsky
* Unknwon https://tiwtter.com/joe2010xtmf

# Licensing

Wharf is licensed under the MIT License.

# We Are Working On Other Projects of Wharf Related

[Vessel](https://githbu.com/dockercn/vessel) A continuous integration system build with Docker.

[Rudder](https://github.com/dockercn/rudder) A Docker client of Golang.
