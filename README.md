<picture>
  <source srcset="https://github.com/user-attachments/assets/a91f6bab-2ad5-4235-916b-2f8b27e30f24" media="(prefers-color-scheme: light)"/>
  <source srcset="https://github.com/user-attachments/assets/f5a532ad-73d3-4935-a27b-b4f20da24778"  media="(prefers-color-scheme: dark)"/>
  <img src="https://github.com/user-attachments/assets/a91f6bab-2ad5-4235-916b-2f8b27e30f24" alt="scan24"/>
</picture>

# Scan24
Service that scans some basic information about the web page.

## Screenshots

<picture>
  <source srcset="https://github.com/user-attachments/assets/6aa14e07-6efe-4a79-bda5-a4a8cfc1a36b" media="(prefers-color-scheme: light)"/>
  <source srcset="https://github.com/user-attachments/assets/6ef1dc8d-04e8-4e32-9ead-b85e4faa0f20"  media="(prefers-color-scheme: dark)"/>
  <img src="https://github.com/user-attachments/assets/6aa14e07-6efe-4a79-bda5-a4a8cfc1a36b" alt="scan24 shows a scan of mysh.dev site"/>
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
