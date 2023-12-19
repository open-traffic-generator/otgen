SHELL = /bin/bash

.PHONY: all
all: install network deploy run

.PHONY: clean
clean: remove-lab network-clean

.PHONY: clean-all
clean-all: clean install-clean

###############################
# Install components
###############################

.PHONY: install
install: install-docker-compose

install-docker-compose: /usr/local/bin/docker-compose
/usr/local/bin/docker-compose:
	sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$$(uname -s)-$$(uname -m)" -o /usr/local/bin/docker-compose
	sudo chmod +x /usr/local/bin/docker-compose

install-clean:
	-sudo rm -f `command -v docker-compose`

###############################
# Test network
###############################

.PHONY: network
network:veth0

veth0: /sys/class/net/veth0
/sys/class/net/veth0:
	sudo ip link add name veth0 type veth peer name veth1
	sudo ip link set dev veth0 up
	sudo ip link set dev veth1 up
	sudo sysctl net.ipv6.conf.veth0.disable_ipv6=1
	sudo sysctl net.ipv6.conf.veth1.disable_ipv6=1

network-mtu:
	sudo ip link set veth0 mtu 9500
	sudo ip link set veth1 mtu 9500

network-clean: veth0-clean
veth0-clean:
	-sudo ip link del name veth0 type veth peer name veth1

###############################
# Deploy lab
###############################

.PHONY: deploy
deploy: deploy-lab

deploy-lab:
	sudo docker-compose up -d

remove-lab:
	sudo docker-compose down


###############################
# Run tests
###############################

.PHONY: run
run: test1 test2 test3 test4 test5 test6

test1:
	@echo RUN with STDIN YAML
	cat otg.yml | otgen --log debug run -k

test2:
	@echo RUN with --api
	cat otg.yml | otgen --log debug run -k --api https://127.0.0.1:8443

test3:
	@echo RUN with --file --yaml
	otgen --log debug run -k --file otg.yml --yaml

test4:
	@echo RUN with --interval
	otgen --log debug run -k --file otg.yml --interval 2s

test5:
	@echo RUN with --metrics flow
	otgen --log debug run -k --file otg.yml --metrics flow

test6:
	@echo RUN with --xeta 1.5
	otgen --log debug run -k --file otg.yml --xeta 1.5
