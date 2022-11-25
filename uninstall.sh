#!/bin/bash

unit_path="/etc/systemd/system/dvnet.service"
exec_path="/usr/local/bin/dvnet"

if [ $EUID -ne 0 ]
then
	printf "The uninstallation should be run as root...\n"
	exit -1
fi

# Stop dvnet before attempting to uninstall it
printf "Stopping dvnet in case it were still running..."
systemctl stop dvnet &> /dev/null
printf " done!\n"

# Remove both the unit file and the executable
printf "Removing the SystemD unit file and the executable..."
rm -f $unit_path $exec_path
printf " done!\n"

# Reload SystemD's config manager to clean it up
printf "Reloading SystemD's manager configuration..."
systemctl daemon-reload
printf " done!\n"

# Done!
printf "Dvnet has been correctly uninstalled! Thanks for giving it a go :P\n"
exit 0
