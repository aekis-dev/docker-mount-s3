docker image build --platform linux/arm64 -t aekis/docker-mount-s3:v1.15.1 -f src/source.Dockerfile .
sudo rm -R src/rootfs
mkdir -p src/rootfs
docker export "$(docker create aekis/docker-mount-s3:v1.15.1 true)" | tar -x -C src/rootfs
docker plugin disable aekis/docker-mount-s3:v1.15.1
docker plugin rm aekis/docker-mount-s3:v1.15.1
docker plugin create aekis/docker-mount-s3:v1.15.1 src/
docker plugin enable aekis/docker-mount-s3:v1.15.1
docker plugin push aekis/docker-mount-s3:v1.15.1
docker plugin install --alias docker-mount-s3:v1.15.1 aekis/docker-mount-s3:v1.15.1
