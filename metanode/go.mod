module com.mgface.disobj/metanode

go 1.15

require (
	com.mgface.disobj/common v0.0.0-00010101000000-000000000000
	github.com/hashicorp/memberlist v0.2.2
	github.com/pborman/uuid v1.2.1
	github.com/sirupsen/logrus v1.7.0
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/timest/env v0.0.0-20180717050204-5fce78d35255
	github.com/urfave/cli v1.22.5
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace com.mgface.disobj/common => ../common
