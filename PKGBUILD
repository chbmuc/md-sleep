pkgname=md-sleep
pkgver=0.0.1
pkgrel=1
pkgdesc='md-sleep watch md-raid array and spin down idle disks'
arch=('x86_64')
url="https://github.com/chbmuc/$pkgname"
license=('MIT')
source=("$url/blob/master/$pkgname.go"
        "md-sleep.service"
        "md-sleep.conf")
noextract=('$pkgname.go')
sha256sums=('2398f330f15c2311f1db707df97080eb4954cd3c30b92d821e7e221e867d32e7'
            '549c89978805778c2ad563d39a79aed41e58dab5c7b2a16f5ee573a2d1196303'
            'SKIP')
install="$pkgname.install"

build() {
  go build \
    -gcflags "all=-trimpath=$PWD" \
    -asmflags "all=-trimpath=$PWD" \
    -ldflags "-extldflags $LDFLAGS" \
    -o $pkgname .
}

package() {
  install -Dm755 $pkgname -t "$pkgdir"/usr/bin/

  install -Dm644 "$srcdir/md-sleep.service" "$pkgdir"/usr/lib/systemd/system/md-sleep.service
  install -Dm644 "$srcdir/md-sleep.conf" "$pkgdir"/etc/md-sleep.conf
}
