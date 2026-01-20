module github.com/username/progetto/gateway-service

go 1.25

toolchain go1.25.0

replace github.com/username/progetto/proto => ../../shared/proto

require (
	github.com/danielgtaylor/huma/v2 v2.34.1
	github.com/go-chi/chi/v5 v5.2.2
)

require (
	github.com/fxamacker/cbor/v2 v2.8.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
)
