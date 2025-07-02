module github.com/kingparks/cursor-vip/auth

go 1.23.0

require (
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/lqqyt2423/go-mitmproxy v1.8.5
)

replace (
	github.com/denisbrodbeck/machineid => ./machineid
	github.com/lqqyt2423/go-mitmproxy => ./go-mitmproxy
)