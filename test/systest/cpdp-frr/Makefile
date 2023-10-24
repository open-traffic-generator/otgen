SHELL = /bin/bash

.PHONY: all
all: install deploy

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
