<div align="center">
  <img src="flakymock.png" alt="FlakyMock Mascot" width="120"/>
</div>



# Quickstart Guide

Get up and running with `flakymock` in minutes.

## 1. Installation

```bash
go get github.com/arun0009/flakymock
# OR
git clone https://github.com/arun0009/flakymock.git
cd flakymock
```

## 2. Run the Server

```bash
go run main.go
```
The server starts on port `8080`.

## 3. Verify It works

Visit `http://localhost:8080/echo` in your browser. You should see a JSON response with your request details.

## 4. Try a Fault

By default, `scenarios.yaml` might be empty or missing. Create one:

```yaml
# scenarios.yaml
- path: /api/hello
  method: GET
  responses:
    - status: 200
      delay: 2s
      body: '{"message": "delayed hello"}'
```

Restart the server and hit `http://localhost:8080/api/hello`. It should take 2 seconds to respond.
