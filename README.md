<picture>
  <source srcset="https://github.com/user-attachments/assets/2b2744a4-a6a1-4abd-a95f-dadd33b0acfb" media="(prefers-color-scheme: light)"/>
  <source srcset="https://github.com/user-attachments/assets/0345b1c5-8e7d-413a-8c25-8811bb47c655"  media="(prefers-color-scheme: dark)"/>
  <img src="https://github.com/user-attachments/assets/2b2744a4-a6a1-4abd-a95f-dadd33b0acfb" alt="scan24 screenshot"/>
</picture>

# Scan24

Service that scans some basic information about the web page.

## Screenshots

<picture>
  <source srcset="https://github.com/user-attachments/assets/d5b2a45c-9980-47ef-977c-21ecd175a516" media="(prefers-color-scheme: light)"/>
  <source srcset="https://github.com/user-attachments/assets/ffa162e4-2ecb-4ee7-a626-ecd98722c4e8"  media="(prefers-color-scheme: dark)"/>
  <img src="https://github.com/user-attachments/assets/d5b2a45c-9980-47ef-977c-21ecd175a516" alt="scan24 shows a scan of mysh.dev site"/>
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

## TODO

- Support for `.env` to make it a little bit more friendly
- Come up with a user-agent and respect robots.txt if possible
- Queue system: If multiple users request the same page concurrently, we should ignore duplicate requests and return the status of the currently running job
- Rate limiting: Scan24 currently spawns a separate goroutine for each valid `<a>` tag on the page, which can result in 1000++ requests to the remote server
  - In addition to potentially functioning as a DoS tool, this behavior could lead to the IP address being blacklisted
- Better error handling for non-standard HTTP statuses, redirect loops, broken streams etc.
- Modularization: Separate the DOCTYPE and Href parsers into packages and refactor them to be more generic
- Custom HTTP client support: Add the ability to inject custom `http.Client` instances into the parsers
- Tests for login form parser, more tests for href and DOCTYPE parsers
