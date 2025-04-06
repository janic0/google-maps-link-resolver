# Google Maps URL Resolver

A very simple dockerized Go app that takes a Google Maps URL from a POST request and returns the coordinates.

## Usage example

```http
POST / HTTP/2
Host: google-maps-link-resolver:8000
Content-Type: text/plain
Accept: application/json
Content-Length: 41

https://maps.app.goo.gl/79CVreRR5Jwg384h7

HTTP/2 200
Content-Type: application/json
Content-Length: 110
```

```json
{
  "ok": true,
  "result": {
    "place_type": "place",
    "lat": 46.6186873,
    "lon": 8.5678655,
    "zoom": "370m",
    "query": "Burgruine"
  }
}
```

## Possible input

- maps.app.goo.gl/\*
- www.google.com/maps*

## Running with docker

```bash
docker run -p 8000:8000 --restart=unless-stopped ghcr.io/janic0/google-maps-link-resolver:latest
```

```yml
services:
  google-maps-url-resolver:
    container_name: google-maps-url-resolver
    image: ghcr.io/janic0/google-maps-link-resolver:latest
    restart: unless-stopped
    ports:
      - 8000:8000
```
