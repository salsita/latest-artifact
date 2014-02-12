# latest-artifact #

This repository contains a tiny web server that is actually not supposed to serve
any content. On the other hand, when hit with a path having its base equal to "latest",
the server returns a redirect to the most recently modified file in the relevant
directory.

## Installation ##

1. Install [Go](http://golang.org/doc/install).
2. Set up a [Go workspace](http://golang.org/doc/code.html#GOPATH).
3. Execute `go get github.com/salsita/latest-artifact`.
4. Set up Nginx or other reverse proxy to cooperate with `latest-artifact`.
5. Run `latest-artifact`, by default it listens on `localhost:9876`.

### Nginx Configuration Example ###

`latest-artifact` is not really supposed to serve the static files themselves.
Nginx can be used for that, and here is an example of how to achive it:

```
server {
        listen 80;
        server_name "artifacts.example.com";

        deny 192.168.23.254;
        allow 192.168.23.0/24;

        access_log /var/log/nginx/artifacts.example.com/http.access.log;
        error_log  /var/log/nginx/artifacts.example.com/http.error.log;

        client_max_body_size 10G;

        location / {
                root "/srv/artifacts/";

                if ($request_method != "GET") {
                        return 405;
                }

                try_files $uri $uri/ @latest;

                autoindex on;
        }

        location @latest {
                proxy_pass "http://127.0.0.1:9876";

                proxy_set_header Host $host;
                proxy_set_header X-Forwarded-Host $host;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Proto $scheme;
                proxy_set_header X-Real-IP $remote_addr;
        }
}

server {
        listen 443 ssl;
        server_name "artifacts.example.com";

        ssl on;
        ssl_certificate /etc/ssl/private/example_com.crt;
        ssl_certificate_key /etc/ssl/private/example_com.key;

        ssl_session_timeout 5m;

        ssl_protocols SSLv3 TLSv1;
        ssl_ciphers ALL:!ADH:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv3:+EXP;
        ssl_prefer_server_ciphers on;

        deny 192.168.23.254;
        allow 192.168.23.0/24;

        access_log /var/log/nginx/artifacts.example.com/https.access.log;
        error_log  /var/log/nginx/artifacts.example.com/https.error.log;

        client_max_body_size 10G;

        location / {
                root "/srv/artifacts/";
                dav_methods "PUT";

                limit_except "GET" {
                        auth_basic "Salsita build artifacts";
                        auth_basic_user_file "/etc/nginx/auth/artifacts_example_com.htpasswd";
                }

                try_files $uri $uri/ @latest;

                autoindex on;
                create_full_put_path on;
        }

        location @latest {
                proxy_pass "http://127.0.0.1:9876";

                proxy_set_header Host $host;
                proxy_set_header X-Forwarded-Host $host;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Proto $scheme;
                proxy_set_header X-Real-IP $remote_addr;
        }
}
```

## License ##

MIT, see the `LICENSE` file.
