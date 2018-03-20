package main

func main() {
	var endpoint = mustSetupEndpoint("localhost:8080")
	endpoint.serve()
}
