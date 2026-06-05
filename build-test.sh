echo "Starting to build debug version"

if [ ! -d ~/.packer.d ]; then
  mkdir ~/.packer.d
fi
if [ ! -d ~/.packer.d/plugins ]; then
  mkdir ~/.packer.d/plugins
fi

echo "Building packer-builder-parallels"
go build -o ~/.packer.d/plugins/packer-builder-parallels
echo "Setting the log environment variable"
export PACKER_LOG=1

