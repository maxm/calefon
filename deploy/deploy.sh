GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o build/main.linux main.go
ssh server <<'ENDSSH'
  mkdir -p /var/www/calefon/
ENDSSH
scp build/main.linux server:/var/www/calefon/main.linux.next
scp deploy/calefon.conf server:/etc/init/
ssh server <<'ENDSSH'
  /sbin/stop calefon
  mv /var/www/calefon/main.linux.next /var/www/calefon/main.linux
  /sbin/start calefon
ENDSSH

