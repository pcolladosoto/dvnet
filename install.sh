#!/bin/bash

# Fail as soon as we encounter an error to avoid broken installations
set -e

version="0.2"

unit_url="https://raw.githubusercontent.com/pcolladosoto/dvnet/main/dvnet.service"
exec_url="https://github.com/pcolladosoto/dvnet/releases/download/v${version}/dvnet"

unit_path="/etc/systemd/system/dvnet.service"
exec_path="/usr/local/bin/dvnet"

if [ $EUID -ne 0 ]
then
	printf "The installation should be run as root...\n"
	exit -1
fi

# Pull the service file
printf "Downloading the SystemD unit file..."
curl -sSL -o $unit_path $unit_url
printf " done!\n"

# Adjust the permissions in case the umask is not the default one
printf "Adjusting the unit file's permissions..."
chmod 0644 $unit_path
printf " done!\n"

# Make SystemD pick up the changes
printf "Reloading SystemD's manager configuration..."
systemctl daemon-reload
printf " done!\n"

# Pull the executable
printf "Downloading the dvnet executable..."
curl -sSL -o $exec_path $exec_url
printf " done!\n"

# Adjust the permissions
printf "Adjusting the executable's permissions..."
chmod 0755 $exec_path
printf " done!\n"

# Done!
printf "%s\n%s\n%s\n\t%s\n\t\t%s\n\t\t%s\n" \
	"Installation done!" \
	"Don't forget to run 'systemctl start dvnet' to kick things off :P" \
	"You can use the following to uninstall dvnet from your machine:" \
	"sudo bash -c 'systemctl stop dvnet &> /dev/null;" \
	"rm -f /usr/local/bin/dvnet /etc/systemd/system/dvnet.service;"\
	"systemctl daemon-reload'"
exit 0
