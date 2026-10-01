# ASCII Art Web

## Description

ASCII Art Web is a simple Go web server that converts text into ASCII art.
This version has a Minecraft theme to it, and some scripting.

It supports these three banners:

- `standard`
- `shadow`
- `thinkertoy`

The project uses Go's standard library only. There is no JavaScript.

## Developer

- Abdullah (aalhejai)
- Tahir (tzakir)

## Design

- Tahir (tzakir)

## Usage: how to run

Open a terminal in the project directory and run:

```bash
go run .
```

Then open:

```text
http://localhost:8080
```

Enter text, choose a banner and press **Render ASCII Art**.

## HTTP endpoints

### GET /

Displays the main page.

Returns:

- `200 OK` for the correct request.
- `404 Not Found` for an unknown path.
- `405 Method Not Allowed` when the wrong HTTP method is used.

### POST /ascii-art

Receives the text and selected banner.

Returns:

- `200 OK` when the request is valid.
- `400 Bad Request` for empty text, text longer than 250 characters, more than 250 lines, Arabic/non-ASCII input, or an invalid banner.
- `405 Method Not Allowed` when `/ascii-art` is requested with GET or another method.
- `404 Not Found` when a requested page, template, or banner file is missing.
- `500 Internal Server Error` for unexpected internal file or rendering errors.

Error responses display a shared page with the status code, message, and a link back to the home page.

## Implementation details: algorithm

1. The server starts on port `8080`.
2. `GET /` loads `templates/index.html`.
3. The form sends a `POST` request to `/ascii-art`.
4. The server checks the input:
   - text must not be empty;
   - maximum 250 characters;
   - maximum 250 lines;
   - only printable ASCII characters and new lines;
   - banner must be `standard`, `shadow`, or `thinkertoy`.
5. The selected banner is loaded from the `banners` directory.
6. Each banner contains 95 ASCII characters. Each character uses 8 rows plus one separator row, giving 855 lines.
7. The program calculates the correct 8 rows for each input character and builds the final ASCII-art result.
8. The result is displayed on the page.

### Banner file protection

The original banner files are checked with SHA-256 hashes when the server starts.

If an existing banner file has been modified, the program stops with a terminal error.

If a banner file is missing, the server can still start. A request that needs the missing banner returns:

```text
404 Not Found
```

with a 404 error page.

This prevents silent use of altered banner data while keeping missing resources as proper HTTP 404 responses.
