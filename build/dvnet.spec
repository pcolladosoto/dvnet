Name:		dvnet
Version:	0.2
Release:	1%{?dist}
Summary:	Docker-based Virtual Networks
BuildArch:	x86_64

License:	GPLv3

Requires:	docker-ce docker-ce-cli containerd.io

BuildRequires:	systemd

# Longer description on what the package is/does
%description
Dvnet implements a Docker plugin capable of parsing ad-hoc network
definitions to them implement them as a set of connected containers.

The daemon can be controlled through the installed dvnet SystemD unit.

Be sure to check https://github.com/pcolladosoto/dvnet for more information.

# Time to copy the binary file!
%install
# Delete the previos build root
rm -rf %{buildroot}

# Create the necessary directories
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_unitdir}

# And install the file
install -m 0775 %{_sourcedir}/bins/%{name}-%{version} %{buildroot}%{_bindir}/%{name}
install -m 0664 %{_sourcedir}/units/%{name}-%{version}.service %{buildroot}%{_unitdir}/%{name}.service

# Files provided by the package
%files
%{_bindir}/%{name}
%{_unitdir}/%{name}.service

# Changes introducd with each version
%changelog
* Sun Oct 22 2023 Pablo Collado Soto <pcolladosoto@gmx.com>
- First RPM-packaged version
