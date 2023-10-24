test:
	go test -cover -coverprofile=coverprofile -v ./...



.ONE_SHELL:
citest:
	apt-get update
	DEBIAN_FRONTEND=noninteractive \
	apt-get install -y \
	  libgstreamer1.0-dev libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev \
	  libgstreamer-plugins-bad1.0-dev gstreamer1.0-plugins-base gstreamer1.0-plugins-good \
	  gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly gstreamer1.0-libav gstreamer1.0-tools \
	  gstreamer1.0-x gstreamer1.0-alsa gstreamer1.0-gl gstreamer1.0-gtk3 gstreamer1.0-qt5 \
	  gstreamer1.0-pulseaudio
	jackd -r -d dummy &>/dev/null &
	sleep 1
	$(MAKE) test


test-ci-test:
	podman build -t test/goci:1.20 test-files/docker/
	podman run --rm -it \
		-v $(shell pwd):/app \
		-w /app \
		test/goci:1.20 \
		make citest


