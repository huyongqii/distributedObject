module com.mgface.disobj/apinode

go 1.15

require (
	com.mgface.disobj/common v0.0.0-00010101000000-000000000000
	github.com/klauspost/cpuid v1.2.1 // indirect
	github.com/klauspost/reedsolomon v1.9.2
	github.com/sirupsen/logrus v1.7.0
	github.com/timest/env v0.0.0-20180717050204-5fce78d35255
	github.com/urfave/cli v1.22.5
)

replace com.mgface.disobj/common => ../common
