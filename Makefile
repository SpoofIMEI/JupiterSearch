build:
	go mod tidy 

	go build -o JupiterServer ./cmd/JupiterServer/main.go
	go build -o JupiterNode ./cmd/JupiterNode/main.go
	go build -o JupiterClient ./cmd/JupiterClient/main.go

	chmod +x JupiterServer JupiterNode JupiterClient

install:
	@echo "\e[0;31mNOTE: If you're getting errors, make sure you run make as root and have go installed!\e[0m"

	go mod tidy

	go build -o /usr/local/bin/JupiterServer ./cmd/JupiterServer/main.go
	go build -o /usr/local/bin/JupiterNode ./cmd/JupiterNode/main.go
	go build -o /usr/local/bin/JupiterClient ./cmd/JupiterClient/main.go

	mkdir -p /etc/JupiterSearch

	id -u JupiterServer >/dev/null 2>&1 || useradd JupiterServer
	id -u JupiterNode >/dev/null 2>&1   || useradd JupiterNode

	cp configs/JupiterNode.service   /etc/systemd/system/
	cp configs/JupiterServer.service /etc/systemd/system/
	cp configs/JupiterNode.conf /etc/JupiterSearch/
	cp configs/JupiterServer.conf /etc/JupiterSearch/
	cp configs/tokenization_regex /etc/JupiterSearch/

	chmod 600 /etc/systemd/system/JupiterNode.service
	chmod 600 /etc/systemd/system/JupiterServer.service

	@echo "\e[0;32mJupiterSearch installed successfully!\e[0m"


uninstall:
	@echo "\e[0;31mNOTE: If you're getting errors, make sure you run make as root and have go installed!\e[0m"

	systemctl disable JupiterServer 
	systemctl disable JupiterNode 
	systemctl stop JupiterServer 
	systemctl stop JupiterNode

	rm -f /usr/local/bin/JupiterServer /usr/local/bin/JupiterNode  /usr/local/bin/JupiterClient 
	rm -rf /etc/JupiterSearch

	if id -u JupiterServer >/dev/null 2>&1; then userdel -f JupiterServer; fi
	if id -u JupiterNode >/dev/null 2>&1;   then userdel -f JupiterNode;   fi

	rm -f /etc/systemd/system/JupiterNode.service 
	rm -f /etc/systemd/system/JupiterServer.service

	@echo "\e[0;32mJupiterSearch uninstalled successfully!\e[0m"