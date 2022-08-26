# vnc_install

## 介绍
[oms](https://github.com/ssbeatty/oms)的插件示例

``` bash
goreleaser --snapshot --skip-publish --rm-dist
```

## 怎样离线安装apt的包？

### ubuntu
参考[wine](https://wiki.winehq.org/Ubuntu_zhcn)

```shell
sudo apt-get clean
sudo apt-get --download-only install lightdm x11vnc
cp -R /var/cache/apt/archives/ /media/deb-pkgs/

cd /media/deb-pkgs/
sudo dpkg -i *.deb
```

### centos

```shell
yum -y install yum-utils

mkdir /media/yum-pkgs/ && cd /media/yum-pkgs/
repotrack x11vnc

rpm -Uvh --force --nodeps *.rpm
```