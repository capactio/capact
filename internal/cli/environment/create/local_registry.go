package create

func CreateLocalRegistry() {
	docker container run -d --name registry.localhost -v local_registry:/var/lib/registry --restart always -p 5000:5000 registry:2
}
