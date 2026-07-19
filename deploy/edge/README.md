# Shared edge connector

This is the only Cloudflare Tunnel connector in the repository. It creates the
external Docker network `realtime-me-edge`; the Status and Library release units
join that network only through route-allowlisting Caddy sidecars. Business API
containers never join the edge network.

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
| Status public API hostname | `http://status-public:8080` |
| Library public API hostname | `http://library-public:8080` |

End the ingress list with the managed `http_status:404` catch-all. Do not create
a Tunnel hostname for Console, Manager owner routes, Status private procedures,
or Library private routes. The two public sidecars strip cookies, authorization,
and the internal management key before forwarding only allowlisted routes.

The public Site Worker consumes the two public hostnames. Console reaches Status
and Library over the hosts' LAN and reaches Manager over loopback. Manager's
device endpoint remains private LAN/OpenVPN ingress protected by mTLS.

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
