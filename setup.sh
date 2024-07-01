git clone https://github.com/axboe/liburing.git
pushd liburing
./configure --cc=gcc --cxx=g++
make -j$(nproc)
sudo make install
popd
rm -rf liburing
