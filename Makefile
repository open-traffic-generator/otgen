all: get build tests

get:
	go get
	go mod tidy

build:
	go build -ldflags="-X 'github.com/open-traffic-generator/otgen/cmd.version=v0.0.0-${USER}'"

tests: tests-create tests-create-devices-flow

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
	./otgen add flow --tx otg1 --rx otg2 --swap | \
	diff test/create/flow-device.swap.yml -

	OTG_FLOW_SMAC_P1="02:11:11:00:01:aa" OTG_FLOW_DMAC_P1="02:11:11:00:02:aa" \
	./otgen create device -n otg1 -p p1  | \
	./otgen add    device -n otg2 -p p2  | \
	./otgen add flow --tx otg1 --rx otg2 | \
	diff test/create/flow-device.mac.env.yml -

	./otgen create device -n otg1 -p p1  --mac "02:11:11:00:01:aa" | \
	./otgen add    device -n otg2 -p p2  --mac "02:11:11:00:02:aa" | \
	./otgen add flow --tx otg1 --rx otg2 | \
	diff test/create/flow-device.mac.yml -

	./otgen create device -n otg1 -p p1  --mac "02:11:11:00:01:aa" | \
	./otgen add    device -n otg2 -p p2  --mac "02:11:11:00:02:aa" | \
	./otgen add flow --tx otg1 --rx otg2 --swap | \
	diff test/create/flow-device.mac.swap.yml -
	@echo
