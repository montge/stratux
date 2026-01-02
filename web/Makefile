ifeq ($(statuxBuild),)
stratuxBuild=`git log -n 1 --pretty=%H`
buildtime=`date +%s`
endif

ifeq ($(STRATUX_HOME),)
	STRATUX_HOME := /opt/stratux
endif

www=$(STRATUX_HOME)/www

all:
	mkdir -p $(www)
	cp -r css $(www)/
	cp -r js $(www)/
	cp -r img $(www)/
	cp -r maui $(www)/
	cp -r plates $(www)/
	cp index.html $(www)/
	cp alert.wav $(www)/
