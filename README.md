# prometheus-showq-exporter

Simple Prometheus exporter for Postfix3's showq socket

## Purpose

This basic Prometheus exporter is aimed at monitoring Postfix' queue using the showq socket (for instance `/var/spool/postfix/public/showq` in Debian).  
This works for showq released with Postfix 3.x as it uses its binary format.

## Install

Having a working Golang environment:

```bash
go install github.com/trazfr/prometheus-showq-exporter@latest
```

## Use

This program is configured through a JSON configuration file.

To run, just `prometheus-showq-exporter config.json`

## Examples of configuration file

This configuration shows:

- A timeout of 5 seconds for each requests (default=10)
- Listens to the port `9092` (default value=`:9091`)
- Postfix `showq` socket available in `/var/spool/postfix/public/showq` (default value)

```json
{
    "timeout": 5,
    "showq_path": "/var/spool/postfix/public/showq",
    "listen": ":9092"
}
```
