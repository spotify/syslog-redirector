# rpmbuild -bb --buildroot=`pwd`/_rpmbuild/ syslog-redirector.spec

%define _topdir    %(pwd)
%define _rpmdir    %{_topdir}
%define _specdir   %{_topdir}
%define _srcrpmdir %{_topdir}
%define _builddir  %{_topdir}

%define _rpmfilename %%{NAME}-%%{VERSION}-%%{RELEASE}.%%{ARCH}.rpm

Name:           syslog-redirector
Version:        %(cat VERSION)
Release:        el7.centos
Summary:        %{name}

Group:          Development/Libraries
License:        Apache License
BuildRoot:      %{_topdir}/_rpmbuild
BuildArch:      %(uname -m)
#Requires:

Prefix: /usr/local

%description

%prep

%build

%install
mkdir -p $RPM_BUILD_ROOT/usr/local/bin
cp -prf syslog-redirector $RPM_BUILD_ROOT/usr/local/bin/syslog-redirector

%clean
rm -fr $RPM_BUILD_ROOT

%pre

%preun

%post

%postun

%files
%defattr(-,root,root,0755)
%attr(755,root,root) /usr/local/bin/syslog-redirector
