# Satisfactory Dedicated Server Exporter

## Don't Panic
> [!IMPORTANT]
> If the exporter shows no metrics you might want to check the [SDSE_INSECURE](#sdse_insecure) environment variable.

## Usage
### docker run
```shell
docker run -p 8080:8080 -e SDSE_HOST=ip_or_hostname -e SDSE_PORT=7777 -e SDSE_TOKEN=token_from_satisfactory_dedicated_server ghcr.io/redstonecrafter0/satisfactory-dedicated-server-exporter:latest
```

### docker compose
```yaml
services:
  satisfactory-dedicated-server-exporter:
    image: ghcr.io/redstonecrafter0/satisfactory-dedicated-server-exporter:latest
    ports:
      - 8080:8080
    environment:
      - SDSE_HOST=ip_or_hostname
      - SDSE_INSECURE=1
      - SDSE_PORT=7777
      - SDSE_TOKEN=token_from_satisfactory_dedicated_server
```

## Variables

### SDSE_HOST
This environment variable contains the hostname or ip of the Satisfactory Dedicated Server you want to monitor.

### SDSE_PORT
This environment variable contains the port of the Satisfactory Dedicated Server you want to monitor.

### SDSE_TOKEN
This environment variable contains the bearer token of the Satisfactory Dedicated Server you want to monitor.
> [!TIP]
> You can use `server.GenerateAPIToken` on the server console to generate this token.

### SDSE_INSECURE
If you have a trusted certificate installed on your Satisfactory Dedicated Server you can ignore this variable.
Otherwise, to ignore certificate verification you need to set this to `1`.

> [!CAUTION]
> You should always prefer using a trusted certificate over ignoring the trust-chain to verify the server.
> You should then also put a TLS terminating reverse proxy in front of this exporter.
