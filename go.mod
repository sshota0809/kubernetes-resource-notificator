module github.com/sshota0809/kubernetes-resource-notificator

go 1.15

replace github.com/go-check/check => github.com/go-check/check v0.0.0-20180628173108-788fd7840127

require (
	github.com/argoproj/notifications-engine v0.2.0
	github.com/spf13/cobra v1.1.3
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
)
