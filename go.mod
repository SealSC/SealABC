module SealABC

go 1.14

replace SealEVM => github.com/AKACoder/SealEVM v0.0.0-20200528104111-c64d3e3ff388

require (
	SealEVM v0.0.0-20200528104111-c64d3e3ff388rmy
	github.com/cbergoon/merkletree v0.2.0
	github.com/gin-gonic/gin v1.5.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.9.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/syndtr/goleveldb v1.0.1-0.20190923125748-758128399b1d
	github.com/urfave/cli/v2 v2.2.0
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
)
