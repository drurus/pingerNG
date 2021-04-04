module drurus/pingerng

go 1.16

replace drurus/drivedb => ./drivedb

require (
	drurus/drivedb v0.0.0-00010101000000-000000000000
	drurus/drivefile v0.0.0-00010101000000-000000000000
	github.com/go-redis/redis/v8 v8.8.0 // indirect
)

replace drurus/drivefile => ./drivefile
