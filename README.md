# Rickroll

A small Go web app that redirects every visitor to Rick Astley's "Never Gonna Give You Up".
Link previews in WhatsApp, Telegram, Discord, iMessage, Slack and other chat apps show
only a harmless title. YouTube never appears there.

## Running

The app runs in Docker. Use the prebuilt image:

```sh
docker run -d --name rickroll -p 8080:8080 ghcr.io/frizzay/rickroll:latest
```

Or build the image yourself:

```sh
docker build -t rickroll .
docker run -d --name rickroll -p 8080:8080 rickroll
```

The server then listens on http://localhost:8080. Any path works, e.g.
`/photos/summer-2026`. Only `/r` is reserved.

To use a different host port, change only the left side of `-p`, e.g. `-p 80:8080`.

## Customizing the preview

The preview is configured with environment variables, passed to `docker run` with `-e`:

| Environment variable  | Default                                               |
|-----------------------|-------------------------------------------------------|
| `PREVIEW_TITLE`       | `Weekend photos 📸`                                   |
| `PREVIEW_DESCRIPTION` | `Here are the pictures I promised you.`               |
| `PREVIEW_IMAGE`       | empty (absolute image URL, e.g. `https://…/image.jpg`) |
| `PREVIEW_SITE_NAME`   | empty                                                 |

```sh
docker run -d --name rickroll -p 8080:8080 \
  -e PREVIEW_TITLE="Team meeting minutes" \
  -e PREVIEW_DESCRIPTION="Please read by Friday" \
  ghcr.io/frizzay/rickroll:latest
```

Environment variables are fixed when the container is created. To change the preview,
remove the container with `docker rm -f rickroll` and start it again with the new values.

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

## Container image

The image builds a static Go binary on `golang:1.27-alpine3.24` and runs on
`alpine:3.24` as the user `nobody`.

The workflow [.github/workflows/container.yml](.github/workflows/container.yml) builds the
image on every push to `main`, on pull requests, on manual dispatch and once a week.
[Trivy](https://trivy.dev) then scans the image. If Trivy finds a vulnerability with
severity `HIGH` or `CRITICAL`, the build fails.

If the workflow runs on `main` and the scan passes, the image is pushed to
`ghcr.io/frizzay/rickroll`, tagged `latest` and with the commit SHA.
Pull requests are only built and scanned, not pushed.
