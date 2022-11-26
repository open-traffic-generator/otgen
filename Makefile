all: get build tests

get:
	go get
	go mod tidy

build:
	go build -ldflags="-X 'github.com/open-traffic-generator/otgen/cmd.version=v0.0.0-${USER}'"

tests: tests-create tests-create-devices-flow tests-add-bgp

tests-create: tests-create-flow-raw

tests-create-flow-raw:
	@echo "#################################################################"
	@echo "# Create raw flow"
	@echo "#################################################################"
	./otgen create flow | diff test/create/flow.defaults.yml -
	./otgen create flow --swap | diff test/create/flow.swap.yml -
	OTG_FLOW_SMAC_P1="02:11:11:00:01:aa" OTG_FLOW_DMAC_P1="02:11:11:00:02:aa" ./otgen create flow | diff test/create/flow.mac.yml -
	./otgen create flow --smac "02:11:11:00:01:aa" --dmac "02:11:11:00:02:aa" | diff test/create/flow.mac.yml -
	./otgen create flow --smac "02:11:11:00:01:aa" --dmac "02:11:11:00:02:aa" --swap | diff test/create/flow.mac.swap.yml -
	@echo

tests-create-devices-flow:
	@echo "#################################################################"
	@echo "# Create two devices with flow between them"
	@echo "#################################################################"
	./otgen create device -n otg1 -p p1  | \
	./otgen add    device -n otg2 -p p2  | \
	./otgen add flow --tx otg1 --rx otg2 | \
	diff test/create/flow-device.defaults.yml -

	./otgen create device -n otg1 -p p1  | \
	./otgen add    device -n otg2 -p p2  | \
	./otgen add flow --tx otg2 --rx otg1 --swap | \
	diff test/create/flow-device.swap.yml -

	OTG_FLOW_SMAC_P1="02:11:11:00:01:aa" ./otgen create device -n otg1 -p p1  | \
	OTG_FLOW_SMAC_P2="02:11:11:00:02:aa" ./otgen add    device -n otg2 -p p2  | \
	./otgen add flow --tx otg1 --rx otg2 | \
	diff test/create/flow-device.mac.env.yml -

	./otgen create device -n otg1 -p p1  --mac "02:11:11:00:01:aa" | \
	./otgen add    device -n otg2 -p p2  --mac "02:11:11:00:02:aa" | \
	./otgen add flow --tx otg1 --rx otg2 | \
	diff test/create/flow-device.mac.yml -

	./otgen create device -n otg1 -p p1  --mac "02:11:11:00:01:aa" | \
	./otgen add    device -n otg2 -p p2  --mac "02:11:11:00:02:aa" | \
	./otgen add flow --tx otg2 --rx otg1 --swap | \
	diff test/create/flow-device.mac.swap.yml -

	./otgen create device -n otg1 -p p1  | \
	./otgen add    device -n otg2 -p p2  | \
	./otgen add flow --tx otg1 --rx otg2 --smac "02:11:11:00:01:aa" --dmac "02:11:11:00:02:aa" | \
	diff test/create/flow-device.flow.mac.yml -

	./otgen create device -n otg1 -p p1 --ip 192.0.2.1 --gw 192.0.2.2 --prefix 30 --location "localhost:5555+localhost:50071" | \
	./otgen add    device -n otg2 -p p2 --ip 192.0.2.5 --gw 192.0.2.6 --prefix 30 --location "localhost:5556+localhost:50072" | \
	./otgen add flow --tx otg1 --rx otg2 --dmac auto --src 192.0.2.1 --dst 192.0.2.5 | \
	diff test/create/flow-device.mac.auto.yml -

	@echo

tests-add-bgp:
	@echo "#################################################################"
	@echo "# Add BGP configuration to a device"
	@echo "#################################################################"
	./otgen create device --name r1 | \
	./otgen --log debug add bgp || \
	echo "Passed"

	./otgen create device | \
	./otgen --log debug add bgp | \
	diff test/add/bgp-device.defaults.yml -

	./otgen create device --name r1 | \
	./otgen --log debug add bgp --device r1 | \
	diff test/add/bgp-device.name.yml -

	./otgen create device | \
	./otgen --log debug add bgp --id 1111 && echo "Expected to fail" && exit 1 || echo Passed

	./otgen create device | \
	./otgen --log debug add bgp --id 1.1.1.1 | \
	diff test/add/bgp-device.id.yml -

	./otgen create device | \
	./otgen --log debug add bgp --asn 1111 | \
	diff test/add/bgp-device.asn.yml -

	./otgen create device | \
	./otgen --log debug add bgp --peer 192.0.2.200 | \
	diff test/add/bgp-device.peer.yml -

	./otgen create device | \
	./otgen --log debug add bgp --peer 192.0.2.200 --type wrong && echo "Expected to fail" && exit 1 || echo Passed

	./otgen create device | \
	./otgen --log debug add bgp --peer 192.0.2.200 --type ibgp | \
	diff test/add/bgp-device.type.yml -

	./otgen create device | \
	./otgen --log debug add bgp --route 198.51.100.0 || \
	echo "Passed"

	./otgen create device | \
	./otgen --log debug add bgp --route /24 || \
	echo "Passed"

	./otgen create device | \
	./otgen --log debug add bgp --route 198.51.100.0/33 || \
	echo "Passed"

	./otgen create device | \
	./otgen --log debug add bgp --route 198.51.100.256/32 || \
	echo "Passed"

	./otgen create device | \
	./otgen --log debug add bgp --route 198.51.100.0/24 | \
	diff test/add/bgp-device.route.yml -

	@echo
