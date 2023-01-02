# compile for version

tar --help > /dev/null
if [ $? -ne 0 ]; then
    echo "tar error"
    exit 1
fi

make
if [ $? -ne 0 ]; then
    echo "make error"
    exit 1
fi

nhole_version=`./bin/nhole-server --version`
echo "build version: $nhole_version"

# cross_compiles
make -f ./Makefile.cross-compiles

rm -rf ./release/packages
mkdir -p ./release/packages

os_all='linux windows darwin freebsd'
arch_all='386 amd64 arm arm64 mips64 mips64le mips mipsle riscv64'

cd ./release

for os in $os_all; do
    for arch in $arch_all; do
        nhole_dir_name="nhole_${nhole_version}_${os}_${arch}"
        nhole_path="./packages/nhole_${nhole_version}_${os}_${arch}"

        if [ "x${os}" = x"windows" ]; then
            if [ ! -f "./nhole-client_${os}_${arch}.exe" ]; then
                continue
            fi
            if [ ! -f "./nhole-server_${os}_${arch}.exe" ]; then
                continue
            fi
            mkdir ${nhole_path}
            mv ./nhole-client_${os}_${arch}.exe ${nhole_path}/nhole-client.exe
            mv ./nhole-server_${os}_${arch}.exe ${nhole_path}/nhole-server.exe
        else
            if [ ! -f "./nhole-client_${os}_${arch}" ]; then
                continue
            fi
            if [ ! -f "./nhole-server_${os}_${arch}" ]; then
                continue
            fi
            mkdir ${nhole_path}
            mv ./nhole-client_${os}_${arch} ${nhole_path}/nhole-client
            mv ./nhole-server_${os}_${arch} ${nhole_path}/nhole-server
        fi
        cp ../LICENSE ${nhole_path}
        cp -rf ../configfiles/* ${nhole_path}

        # packages
        cd ./packages
        if [ "x${os}" = x"windows" ]; then
#            zip -rq ${nhole_dir_name}.zip ${nhole_dir_name}
            tar -zcf ${nhole_dir_name}.tar.gz ${nhole_dir_name}
        else
            tar -zcf ${nhole_dir_name}.tar.gz ${nhole_dir_name}
        fi
        cd ..
        rm -rf ${nhole_path}
    done
done

cd -
