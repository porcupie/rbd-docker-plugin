# spec file for package ceph
#
# Copyright (C) 2015-2016 XTAO technology <rtlinux@163.com>
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon.
#
# This file is under GPL License

%define _topdir %(echo $PWD/..)/

Name:	rbd-docker-plugin           
Version: @VERSION@
Release:        1%{?dist}
Summary:  docker volume plugin based on ceph rbd      

License:  GPL      
URL:    www.xtao.com        
Source0: %{name}.tar.bz2      

BuildRequires:  gcc
BuildRequires:  golang >= 1.7.4

Provides:  rados(github.com/ceph/go-ceph/rados)
Provides:  rbd(github.com/ceph/go-ceph/rbd)
Provides:  volume(github.com/docker/go-plugins-helpers/volume)

%description
hehe rbd docker plugin

%prep
%setup -q -n rbd-docker-plugin 

# many golang binaries are "vendoring" (bundling) sources, so remove them. Those dependencies need to be packaged independently.
rm -rf vendor

%build
# set up temporary build gopath, and put our directory there
mkdir -p ./_build/src/github.com/
ln -s $(pwd) ./_build/src/github.com/
export GOPATH=$(pwd)/_build:%{gopath}:$HOME
go build -o rbd-docker-plugin .

%install
install -d %{buildroot}%{_bindir}
install -d %{buildroot}/usr/lib/systemd/system/
install -p -m 0755 rbd-docker-plugin %{buildroot}%{_bindir}/rbd-docker-plugin
install -p etc/systemd/rbd-docker-plugin.service %{buildroot}/usr/lib/systemd/system/rbd-docker-plugin.service

%files
%defattr(-,root,root,-)
%doc LICENSE CHANGELOG.md README.md
%{_bindir}/rbd-docker-plugin
/usr/lib/systemd/system/rbd-docker-plugin.service

%changelog
