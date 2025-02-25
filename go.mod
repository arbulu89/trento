module github.com/trento-project/trento

go 1.16

require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/gin-contrib/sessions v0.0.3
	github.com/gin-gonic/gin v1.7.0
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/gomarkdown/markdown v0.0.0-20210514010506-3b9f47219fe7
	github.com/google/uuid v1.3.0
	github.com/hashicorp/consul-template v0.25.2
	github.com/hashicorp/consul/api v1.4.0
	github.com/hooklift/gowsdl v0.5.0
	github.com/lib/pq v1.10.2
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil v3.21.10+incompatible
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.1.2
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/files v0.0.0-20210815190702-a29dd2bc99b2
	github.com/swaggo/gin-swagger v1.3.1
	github.com/swaggo/swag v1.7.4
	github.com/tdewolff/minify/v2 v2.9.16
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/vektra/mockery/v2 v2.9.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/tools v0.1.5 // indirect
	gorm.io/datatypes v1.0.2
	gorm.io/driver/postgres v1.1.2
	gorm.io/gorm v1.21.15
)

replace github.com/trento-project/trento => ./
