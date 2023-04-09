module client

go 1.19

require (
	example.com/common v0.0.0-00010101000000-000000000000
	github.com/gofiber/fiber/v2 v2.38.1
	github.com/streadway/amqp v1.0.0
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.40.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20221013171732-95e765b1cc43 // indirect
)

replace example.com/common => ../common
