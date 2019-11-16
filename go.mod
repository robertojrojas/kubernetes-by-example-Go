module k8s-by-examples

go 1.13

require (
	github.com/imdario/mergo v0.3.8 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	k8s.io/api v0.0.0-20190819141258-3544db3b9e44
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/kind v0.5.1
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77
