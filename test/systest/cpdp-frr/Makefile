SHELL = /bin/bash

.PHONY: all
all: install deploy run

.PHONY: clean
clean: remove-clab

.PHONY: clean-all
clean-all: clean install-clean

###############################
# Install components
###############################

.PHONY: install
install: install-clab

install-clab: /usr/bin/containerlab
/usr/bin/containerlab:
	bash -c "$$(curl -sL https://get.containerlab.dev)" -- -v 0.46.2

install-clean:
	-sudo apt remove containerlab -y

###############################
# Deploy lab
###############################

.PHONY: deploy
deploy: deploy-clab

deploy-clab:
	sudo -E containerlab deploy --reconfigure

remove-clab:
	sudo containerlab destroy --cleanup

###############################
# Run tests
###############################

.PHONY: run
run: test1 test2 test3 test4 test5 test6 test7 test8

test1:
	@echo "CREATE --dmac auto | RUN"
	OTG_LOCATION_P1="localhost:5555+localhost:50071" \
	OTG_LOCATION_P2="localhost:5556+localhost:50072" \
	otgen create device -n otg1 -p p1 --ip 192.0.2.1 --gw 192.0.2.2 --prefix 30 | otgen add device -n otg2 -p p2 --ip 192.0.2.5 --gw 192.0.2.6 --prefix 30 | otgen --log debug add flow --tx otg1 --rx otg2 --dmac auto --src 192.0.2.1 --dst 192.0.2.5 | otgen --log debug run -k --metrics flow,port

test2:
	@echo RUN --rxbgp 2x
	OTG_LOCATION_P1="localhost:5555+localhost:50071" \
	OTG_LOCATION_P2="localhost:5556+localhost:50072" \
	otgen --log debug run -k --file otg.json --json --rxbgp 2x

test3:
	@echo RUN --rxbgp 2x --metrics bgp4
	OTG_LOCATION_P1="localhost:5555+localhost:50071" \
	OTG_LOCATION_P2="localhost:5556+localhost:50072" \
	otgen --log debug run -k --file otg.json --json --rxbgp 2x --metrics bgp4

test4:
	@echo RUN --rxbgp 2x --metrics bgp4,port
	OTG_LOCATION_P1="localhost:5555+localhost:50071" \
	OTG_LOCATION_P2="localhost:5556+localhost:50072" \
	otgen --log debug run -k --file otg.json --json --rxbgp 2x --metrics bgp4,port

test5:
	@echo RUN --rxbgp 2x --metrics bgp4,flow
	OTG_LOCATION_P1="localhost:5555+localhost:50071" \
	OTG_LOCATION_P2="localhost:5556+localhost:50072" \
	otgen --log debug run -k --file otg.json --json --rxbgp 2x --metrics bgp4,flow

test6:
	@echo RUN --rxbgp 2x --metrics bgp4,flow,port
	OTG_LOCATION_P1="localhost:5555+localhost:50071" \
	OTG_LOCATION_P2="localhost:5556+localhost:50072" \
	otgen --log debug run -k --file otg.json --json --rxbgp 2x --metrics bgp4,flow,port

test7:
	@echo RUN --rxbgp 2x --metrics flow,port
	OTG_LOCATION_P1="localhost:5555+localhost:50071" \
	OTG_LOCATION_P2="localhost:5556+localhost:50072" \
	otgen --log debug run -k --file otg.json --json --rxbgp 2x --metrics flow,port

test8:
	@echo "CREATE --dmac auto | ADD BGP | RUN --rxbgp 2x --metrics bgp4,flow"
	OTG_LOCATION_P1="localhost:5555+localhost:50071" \
	otgen create device -n otg1 -p p1 --ip 192.0.2.1 --gw 192.0.2.2 --prefix 30 | \
	otgen add bgp -d otg1 --id 1.1.1.1 --asn 1111 --peer 192.0.2.2 --route 198.51.100.0/24 | \
	OTG_LOCATION_P2="localhost:5556+localhost:50072" \
	otgen add device -n otg2 -p p2 --ip 192.0.2.5 --gw 192.0.2.6 --prefix 30 | \
	otgen add bgp -d otg2 --id 2.2.2.2 --asn 2222 --peer 192.0.2.6 --route 203.0.113.0/24 | \
	otgen add flow --tx otg1 --rx otg2 --dmac auto --src 198.51.100.1 --dst 203.0.113.1 | \
	otgen --log debug run -k --rxbgp 2x --metrics bgp4,flow
