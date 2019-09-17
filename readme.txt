
cd /home/bob/auto/
scp -r root@151.155.216.148:/home/bob/auto/jmon .  This will copy the jmon directory to the current directory.
cd jmon
screen -S jmon
vim /etc/security/limits.conf

	*         hard    nofile      500000
	*         soft    nofile      500000
	root      hard    nofile      500000
	root      soft    nofile      500000