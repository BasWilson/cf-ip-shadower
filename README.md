# What does it do

Checks if the host's public IP has changed from the last time it was checked. If it changes, it will update A records in a Cloudflare zone to point to the new IP.

Useful for dynamic IPs. You can run it on a machine that has a dynamic IP and it will keep the DNS records updated.

## How to use

Compile it with `compile` or `compile_linux` if you are building it for a linux machine from a different OS.

```bash
make compile
```

You can either just run the binary without any input, this will start an http server on port 1338 and will listen for requests on `/` with a POST method.

Or you can run the binary with a path to a json file that contains the records you want to check and update.

Both the body of the request and the json file should have the following format:

```json
{
    "records": [
        {
            "name": "test.domain.com",
            "addr": "127.0.0.1"
        },
        {
            "name": "test2.domain.com"
        }
    ]
}
```

If you provide an `addr` field, it will work with that IP. If you don't provide it, it will use the public IP of the host.

### Example

With a json file called `records.json`:

```bash
./bin/twinner records.json
```

#### With a request:

Start the server:

```bash
./bin/twinner
```

Send a request:

```bash
curl -X POST -d '{"records":[{"name":"test.domain.com","addr":"127.0.0.1"},{"name":"test2.domain.com"}]}' "http://localhost:1338"
```

If you provide a file, it will check if the IP changed every minute. If you run the binary without input, it will only check when you send a request.

You can change the JSON file while the server is running and it will pick up the changes on the next iteration.

The HTTP server does not start if you provide a file.

## Environment variables

```bash
CLOUDFLARE_ZONE_ID=your-zone-id
CLOUDFLARE_API_TOKEN=your-api-token
AUTH_HEADER=your-auth-header-value
```
