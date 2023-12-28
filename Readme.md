
## build

`CGO_ENABLED=0 go build`

and upload it to speedy server:

`scp dummy-go-service non-root-user@your-server:~/`

## run it:

`DUMMY_SERVER_BIND="your-server-ip:8000" nohup ./dummy-go-webservice &`

to run it on lower port than 1024 ... you need to run it as root, which is highly discouraged!

use setcap, or use some iptables magic: https://superuser.com/questions/710253/allow-non-root-process-to-bind-to-port-80-and-443

`sudo setcap CAP_NET_BIND_SERVICE=+eip /path/to/binary`


## example download

(size = megabyte, here 1GB)

`curl -o /dev/null http://your-server:8000/file?size=1000`

whereas speed is in Byte (kB, MB, GB,..)

## example upload

best method:

first get a 100mb file:

`curl -o /home/user/100mb http://your-server:8000/file?size=100`

then upload it ... 

`curl -F 'data=@/home/user/100mb' http://your-server:8000/upload | cat`

whereas speed is in Byte (kB, MB, GB,..) .. so 5000k is like 50Mbit

