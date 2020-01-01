cd /var/www/html/instagram
git pull
cd /var/www/html/instagram/cmd
go build -o ../instagram


/var/www/html/instagram/instagram &> /dev/null &
