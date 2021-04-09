module drurus/pingerng

go 1.16

replace drurus/drivedb => ./drivedb

require (
	drurus/config v0.0.0-00010101000000-000000000000
	drurus/drivedb v0.0.0-00010101000000-000000000000
	drurus/drivefile v0.0.0-00010101000000-000000000000
	drurus/pingtools v0.0.0-00010101000000-000000000000
	github.com/go-ping/ping v0.0.0-20210402232549-1726e5ede5b6 // indirect
	github.com/go-redis/redis/v8 v8.8.0 // indirect
	github.com/joho/godotenv v1.3.0 // indirect
	github.com/labstack/echo/v4 v4.2.1
)

replace drurus/drivefile => ./drivefile

replace drurus/pingtools => ./pingtools

replace drurus/config => ./config
