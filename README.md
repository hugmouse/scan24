<picture>
  <source srcset="https://i.imgur.com/PiBAxUK.png" media="(prefers-color-scheme: light)"/>
  <source srcset="https://i.imgur.com/Tlwi5IF.png"  media="(prefers-color-scheme: dark)"/>
  <img src="https://i.imgur.com/PiBAxUK.png" alt="scan24"/>
</picture>

# Scan24
Service that scans some basic information about the web page.

## Screenshots

<picture>
  <source srcset="https://i.imgur.com/zcwg6eH.png" media="(prefers-color-scheme: light)"/>
  <source srcset="https://i.imgur.com/WN1LCj2.png"  media="(prefers-color-scheme: dark)"/>
  <img src="https://i.imgur.com/zcwg6eH.png" alt="scan24 shows a scan of mysh.dev site"/>
</picture>

## Deploying

By default, Scan24 runs on `:8080`, which means you can access it by navigating to `http://localhost:8080` or `http://YOUR_SERVER_IP:8080`.

### Using Docker

For this one you have to have:

- `Docker` with BuildKit
- `Git` to copy this repo

```bash
git clone https://github.com/hugmouse/scan24.git
cd scan24
docker compose up
```

Then you can navigate in your browser to `http://localhost:8080`

Watch mode is also supported: `docker compose watch`

### Building from source

Make sure you have the following:

- `Go 1.24`
- `Git`

Then we can start building this thing:

```bash
git clone https://github.com/hugmouse/scan24.git
cd scan24
go build cmd/server/main.go -o scan24-server
```

And now you have `scan24-server` executable!

## Hacking

Directory structure

- `cmd/` contains Scan24 server
- `static/` contains reusable static assets, like CSS and SVG
- `templates/` contains Go HTML templates
- `test/` is where all tests go
- `internal/` contains various parsers and HTTP handlers for Scan24

## Related

I used these articles as references when building this service.
If you're interested in parsing `DOCTYPE` and such – check them out.

Wikipedia articles:

- https://en.wikipedia.org/wiki/Document_type_declaration
- https://en.wikipedia.org/wiki/Formal_Public_Identifier
- https://en.wikipedia.org/wiki/Standard_Generalized_Markup_Language

1Password articles:

- https://developer.1password.com/docs/web/compatible-website-design/
