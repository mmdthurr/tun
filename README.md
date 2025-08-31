tun is dead simple tls tunnel previously tlstun built on top of smux.
i have tried to build my own framer and session manager with connection pool along, but the performance was horribly low.

l.conf 
```
{  
 "mode": "listener",  
 "laddr": "0.0.0.0:5050",  
 "sec": "tls",  
 "tls": {  
   "cert": "/home/mmd/code/tun/tst/tls.cert",  
   "key": "/home/mmd/code/tun/tst/tls.key"  
 },  
 "trustedpeers": [  
   "127.0.0.1",  
   "2.3.4.5"  
 ]  
}
```

d.conf
```
{  
 "mode": "dialer",  
 "bckaddr": "192.168.1.7:1086",  
 "peers": [  
   {  
     "addr": "127.0.0.1:5050",  
     "tls": true,  
     "poolsize": 10  
   }  
 ]  
}
```


tun.service
```
[Unit]  
Description=tun  
After=network-online.target  
Wants=network-online.target  
  
[Service]  
ExecStart=/usr/bin/tun -c /etc/tun/config.json  
Restart=always  
RestartSec=0s  
User=root  
  
[Install]  
WantedBy=default.target
```

