# Manifesto

A simple tool to generate a Greybus manifest blob from a Git config style
syntax input file.

Provided under BSD license. See *LICENSE* for details.

## Install

`export GOPATH=.../go`

`go get code.google.com/p/gcfg`

## Run

`go run manifesto.go test.mnfs`

Generates *test.mnfb* in same directory as source file

## Build

To generate a binary:

`go build manifesto.go`

and run as:

`./manifesto test.mnfs`

## Example

### Simple I2C Module
```
;
; Simple I2C Module Manifest
;

[manifest-header]
version-major = 0
version-minor = 1

[module-descriptor]
vendor = 0xdead
product = 0xbeef
version = 1
vendor-string-id = 1
product-string-id = 2
serial-number = 0

; I2C function on CPort 1
[function-descriptor "0"]
cport = 1
function-type = 0x07

; Module vendor string (id can't be 0)
[string-descriptor "0"]
id = 1
string = "Project Ara"

; Module product string (id can't be 0)
[string-descriptor "1"]
id = 2
string = "Simple I2C Module"

; CPort 1
[cport-descriptor "1"]
id = 1
```

### Build Simple I2C Module manifest blob

`go run manifesto.go examples/simple-i2c-module.mnfs`
