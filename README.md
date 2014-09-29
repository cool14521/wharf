docker-bucket
===============

The dockercn's [docker-bucket](https://github.com/dockercn/docker-hub) is a [Golang](http://golang.org) version what clone and beyond official [docker-registry](https://github.com/docker/docker-registry), and we add user manage, UI and more features. We will add more backend storage services support like [Qiniu](http://qiniu.com), [Aliyun OSS](http://www.aliyun.com/product/oss), [Baidu Storage](http://developer.baidu.com/cloud/stor), [Tencent COS](http://www.qcloud.com/product/product.php?item=cos) and [OpenStack Swift](http://docs.openstack.org/developer/swift).

What's Bucket?
================

A bucket is a hosted service containing [repositories](http://docs.docker.io/en/latest/terms/repository/#repository-def) of [images](http://docs.docker.io/en/latest/terms/image/#image-def) which responds to the Registry API.

What's FQIN?
============

A Fully Qualified Image Name (FQIN) can be made up of 3 parts:

```
[registry_hostname[:port]/][user_name/](repository_name[:version_tag])
```

`version_tag` defaults to `latest`, `username` and `registry_hostname` default to an empty string. When `registry_hostname` is an empty string, then docker push will push to *index.docker.io:80*.

Why image file named layer?
===========================

In Docker terminology, a read-only [Layer](http://docs.docker.io/en/latest/terms/layer/#layer-def) is called an image. An image never changes.

LedisDB Key Conf
================

在 LedisDB 中对象的 Key 规则：

```
@Username // @用户名 
#Organization // #组织名

@Username$Repository+ // 用户未加密公有仓库
#Organization$Repository+ // 组织未加密公有仓库

@Username$Repository- // 用户未加密私有仓库
#Organization$Repository- // 组织未加密私公有仓库

@Username$Repository-?Sign // 用户加密私有仓库  
#Organization$Repository-?Sign // 组织加密私有仓库

&Image+ //未加密，私有库和公有库未加密的 Image 共享
&Image-?Sign //加密，只有私有库有加密支持，每个 Image 根据加密签名不同，可能存有多份儿。

@Username$Repository*Template+(-) //  
#Organization$Repository*Template+(-) // 

@Username$Repository!Job //  
#Organization$Repository!Job // 
```

Nginx Conf
==========

```
upstream hub_upstream {
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
    proxy_pass         http://hub_upstream;
  }
}
```
