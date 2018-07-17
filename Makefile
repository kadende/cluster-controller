test:
	go test -v ./...

dependecies:
	dep ensure -v


	#dlv --listen=:2345 --headless=true --api-version=2 test ./plugin-manager/ -- -run ^TestHello TestInstallingProvider