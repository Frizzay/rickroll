# Rickroll

A small Go web app that redirects every visitor to Rick Astley's "Never Gonna Give You Up".
Link previews in WhatsApp, Telegram, Discord, iMessage, Slack and other chat apps show
only a harmless title. YouTube never appears there.

## Running

Requires Go 1.22 or newer.

```sh
go run .
```

The server then listens on http://localhost:8080. Any path works, e.g.
`/photos/summer-2026`. Only `/r` is reserved.

## Customizing the preview

Everything can be set with a flag or an environment variable:

| Flag           | Environment variable  | Default                                               |
|----------------|-----------------------|-------------------------------------------------------|
| `-addr`        | `PORT`                | `:8080`                                               |
| `-title`       | `PREVIEW_TITLE`       | `Weekend photos 📸`                                   |
| `-description` | `PREVIEW_DESCRIPTION` | `Here are the pictures I promised you.`               |
| `-image`       | `PREVIEW_IMAGE`       | empty (absolute image URL, e.g. `https://…/image.jpg`) |
| `-site-name`   | `PREVIEW_SITE_NAME`   | empty                                                 |

```sh
go run . -title "Team meeting minutes" -description "Please read by Friday"
```

## How it works

1. The shared URL **never** answers with an HTTP redirect. Preview bots follow redirects
   and would otherwise show YouTube's title and thumbnail.
2. Preview bots are detected by their user agent and get a page that contains only the
   preview metadata (`og:title`, `og:description`, `og:image`) and no script.
3. Real browsers get the same page plus a script that redirects to `/r`.
   Only `/r` redirects to YouTube with a 302. The HTML therefore never mentions "youtube".

## Notes

- Chat apps build the preview by fetching the URL themselves. For a preview to appear,
  the server must be publicly reachable (e.g. on a VPS, Fly.io or Render, or for testing
  through a tunnel such as `cloudflared` or `ngrok`).
- The domain is of course still visible in the link. An inconspicuous domain name helps.
- Chat apps cache previews. If you change the title or image, test with a new path
  (e.g. `/photos2`).

## Tests

```sh
go test ./...
```

## Container

```sh
docker build -t rickroll .
docker run -p 8080:8080 -e PREVIEW_TITLE="Team meeting minutes" rickroll
```

The image builds a static Go binary on `golang:1.27-alpine3.24` and runs on
`alpine:3.24` as the user `nobody`.

The workflow [.github/workflows/container.yml](.github/workflows/container.yml) builds the
image on every push to `main`, on pull requests, on manual dispatch and once a week.
[Trivy](https://trivy.dev) then scans the image. If Trivy finds a vulnerability with
severity `HIGH` or `CRITICAL`, the build fails.

If the workflow runs on `main` and the scan passes, the image is pushed to
`ghcr.io/frizzay/rickroll`, tagged `latest` and with the commit SHA.
Pull requests are only built and scanned, not pushed.

```sh
docker run -p 8080:8080 ghcr.io/frizzay/rickroll:latest
```
