# Shared edge connector

This is the only Cloudflare Tunnel connector in the repository. It creates the
external Docker network `realtime-me-edge`; the Status and Library release units
join that network through stable aliases while retaining independent Compose
projects and lifecycle controls.

## Configure

Create a remotely managed tunnel, store its token in a root-only file, and copy
the Compose settings:

```sh
sudo install -d -m 0700 /etc/realtime-me
sudo install -m 0400 /path/to/tunnel.token /etc/realtime-me/cloudflare-tunnel.token
cp deploy/edge/.env.example deploy/edge/.env
```

Configure the tunnel's public hostnames in Cloudflare before starting it:

| Public hostname | Origin service |
| --- | --- |
| Status API hostname | `http://status-api:8080` |
| Library private API hostname | `http://library-api:8080` |
| Library public API hostname | `http://library-api:8080` |

End the ingress list with the managed `http_status:404` catch-all. Public and
private Library hosts intentionally reach the same API process; exact Host routing,
OIDC permission checks, and a reduced public router preserve the application boundary.

The public Site Worker consumes the Status and Library public hostnames. The
Console BFF consumes Status and the Library private hostname. Manager remains on
its direct host: mTLS device traffic uses the public Manager endpoint while Console
owner traffic reaches Manager over loopback.

## Run

Create the shared network first, then start the two origin units, and finally
start the connector:

```sh
docker compose --env-file deploy/edge/.env -f deploy/edge/compose.yaml create
docker compose --env-file deploy/status/.env -f deploy/status/compose.yaml up -d --build
sudo /opt/cloud-drive/deploy/library/scripts/deploy.sh
docker compose --env-file deploy/edge/.env -f deploy/edge/compose.yaml up -d --wait
```

Stopping `cloudflared` interrupts only public ingress. Do not run `down` while
Status or Library containers are attached to `realtime-me-edge`; use `stop` or
restart the connector instead.
