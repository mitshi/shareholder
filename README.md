# physicalshare

Simple web service to update our Shareholders email and mobile numbers.

## Develop

First install latest Go version. Ensure it is in your path.

```
cp env.example .env
# Update local .env

# First time install
go install

export $(cat .env | xargs) go run *.go
```

## Deploy

TODO
