# Mantle Container Support

## Command Summary

 - `build-builder` Builds a docker image that contains tools for building mantle.  Default docker image tag is `mantle-builder:${VERSION}${ARCH_TAG}`.
 - `build-mantle` Builds mantle programs using a mantle-builder container.
 - `build-runner` Builds a docker image that containes the mantle programs.  Default docker image tag is `mantle-runner:${VERSION}${ARCH_TAG}`.
 - `run-mantle` Runs mantle programs in a mantle-runner container.

## Examples

### Build the mantle-runner Container Image

    ./build-builder -v
    ./build-mantle -vt
    ./build-runner -v

### Run Kola Tests in a mantle-runner Container

See https://coreos.com/releases for release info.

    CHANNEL=alpha
    BOARD=amd64-usr
    #BOARD=arm64-usr
    RELEASE=current
    SERVER=https://${CHANNEL}.release.core-os.net/${BOARD}/${RELEASE}
    RELEASES=/tmp/coreos-releses

    mkdir -p ${RELEASES}/${CHANNEL}/${BOARD}/${RELEASE}
    pushd ${RELEASES}/${CHANNEL}/${BOARD}/${RELEASE}
    wget --no-verbose ${SERVER}/coreos_production_qemu_uefi{.DIGESTS,.sh,_efi_code.fd,_efi_vars.fd}
    wget --no-verbose ${SERVER}/coreos_production_image.bin.bz2{.DIGESTS,}
    bunzip2 coreos_production_image.bin.bz2
    popd

    ls ${RELEASES}/${CHANNEL}/${BOARD}/${RELEASE}
    DOCKER_ARGS="-v ${RELEASES}:/releases:ro" ./run-mantle -kv -- kola run coreos.basic --board="${BOARD}" --parallel=2 --platform=qemu --qemu-bios=/releases/${CHANNEL}/${BOARD}/${RELEASE}/coreos_production_qemu_uefi_efi_code.fd --qemu-image=/releases/${CHANNEL}/${BOARD}/${RELEASE}/coreos_production_image.bin
